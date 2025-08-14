package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/temidaradev/NeuraTalk/internal"
)

type Server struct {
	upgrader websocket.Upgrader
	clients  map[*websocket.Conn]bool
}

type Message struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type ChatRequest struct {
	Model   string   `json:"model"`
	Message string   `json:"message"`
	History []string `json:"history"`
}

type ChatResponse struct {
	Response string `json:"response"`
	Error    string `json:"error,omitempty"`
}

type ModelsResponse struct {
	Models []string `json:"models"`
	Error  string   `json:"error,omitempty"`
}

func NewServer() *Server {
	return &Server{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for development
			},
		},
		clients: make(map[*websocket.Conn]bool),
	}
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	s.clients[conn] = true
	defer delete(s.clients, conn)

	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			break
		}

		switch msg.Type {
		case "chat_request":
			s.handleChatRequest(conn, msg.Payload)
		case "get_models":
			s.handleGetModels(conn)
		}
	}
}

func (s *Server) handleChatRequest(conn *websocket.Conn, payload interface{}) {
	data, _ := json.Marshal(payload)
	var req ChatRequest
	if err := json.Unmarshal(data, &req); err != nil {
		s.sendError(conn, "Invalid chat request format")
		return
	}

	// Use the existing internal package for Ollama integration
	response, err := s.generateOllamaResponse(req.Model, req.Message, req.History)
	if err != nil {
		s.sendError(conn, fmt.Sprintf("Error generating response: %v", err))
		return
	}

	msg := Message{
		Type: "chat_response",
		Payload: ChatResponse{
			Response: response,
		},
	}

	conn.WriteJSON(msg)
}

func (s *Server) handleGetModels(conn *websocket.Conn) {
	models, err := internal.GetAvailableModels()
	if err != nil {
		s.sendError(conn, fmt.Sprintf("Error getting models: %v", err))
		return
	}

	msg := Message{
		Type: "models_response",
		Payload: ModelsResponse{
			Models: models,
		},
	}

	conn.WriteJSON(msg)
}

func (s *Server) sendError(conn *websocket.Conn, errorMsg string) {
	msg := Message{
		Type: "error",
		Payload: map[string]string{
			"error": errorMsg,
		},
	}
	conn.WriteJSON(msg)
}

func (s *Server) generateOllamaResponse(model, message string, history []string) (string, error) {
	// Create a temporary InputOutput instance for generating responses
	// We'll modify the internal package to support headless mode
	return internal.GenerateOllamaResponseHeadless(model, message, history)
}

// HTTP handlers for REST API (alternative to WebSocket)
func (s *Server) handleModelsHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		return
	}

	models, err := internal.GetAvailableModels()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ModelsResponse{Error: err.Error()})
		return
	}

	json.NewEncoder(w).Encode(ModelsResponse{Models: models})
}

func (s *Server) handleChatHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		return
	}

	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ChatResponse{Error: "Invalid request format"})
		return
	}

	response, err := s.generateOllamaResponse(req.Model, req.Message, req.History)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ChatResponse{Error: err.Error()})
		return
	}

	json.NewEncoder(w).Encode(ChatResponse{Response: response})
}

// File operations for conversation storage
func (s *Server) handleSaveConversation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		return
	}

	type SaveRequest struct {
		Model        string   `json:"model"`
		Conversation []string `json:"conversation"`
	}

	var req SaveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Create conversations directory structure
	modelDir := filepath.Join("./conversations", req.Model)
	os.MkdirAll(modelDir, 0755)

	// Create timestamped file
	timestamp := time.Now().Format("20060102_150405")
	fileName := fmt.Sprintf("%s_%s.txt", req.Model, timestamp)
	filePath := filepath.Join(modelDir, fileName)

	content := strings.Join(req.Conversation, "\n\n")
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleLoadConversation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	model := r.URL.Query().Get("model")
	if model == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	filePath := filepath.Join("./tmp", model+".txt")
	content, err := os.ReadFile(filePath)
	if err != nil {
		// File doesn't exist, return empty conversation
		json.NewEncoder(w).Encode(map[string]interface{}{
			"conversation": []string{},
		})
		return
	}

	conversation := []string{}
	if len(content) > 0 {
		conversations := strings.Split(string(content), "\n\n")
		for _, conv := range conversations {
			if strings.TrimSpace(conv) != "" {
				conversation = append(conversation, conv)
			}
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"conversation": conversation,
	})
}

func main() {
	server := NewServer()

	r := mux.NewRouter()

	// WebSocket endpoint
	r.HandleFunc("/ws", server.handleWebSocket)

	// REST API endpoints
	r.HandleFunc("/api/models", server.handleModelsHTTP).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/chat", server.handleChatHTTP).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/conversation/save", server.handleSaveConversation).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/conversation/load", server.handleLoadConversation).Methods("GET", "OPTIONS")

	// Serve static files for the WASM app with proper MIME types
	fileServer := http.FileServer(http.Dir("./web/"))
	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set correct MIME type for WASM files
		if strings.HasSuffix(r.URL.Path, ".wasm") {
			w.Header().Set("Content-Type", "application/wasm")
		}
		fileServer.ServeHTTP(w, r)
	})

	fmt.Println("Server starting on :8080")
	fmt.Println("WebSocket endpoint: ws://localhost:8080/ws")
	fmt.Println("REST API: http://localhost:8080/api/")
	fmt.Println("Web UI: http://localhost:8080/")

	log.Fatal(http.ListenAndServe(":8080", r))
}
