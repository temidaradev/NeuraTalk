package internal

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

type InputOutput struct {
	InputEntry      *widget.Entry
	OutputLabel     *widget.Label
	ModelSelect     *widget.Select
	SelectedModel   string
	ParentWindow    fyne.Window
	ScrollContainer *container.Scroll
	Conversation    []string
	ClearButton     *widget.Button
	Settings        *Settings

	animating       bool
	animationTicker *time.Ticker
}

func isFileEmpty(filePath string) (bool, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return false, err
	}

	return fileInfo.Size() == 0, nil
}

func ensureTmpDirectoryExists() {
	// Create tmp directory if it doesn't exist
	if _, err := os.Stat("./tmp"); os.IsNotExist(err) {
		err := os.Mkdir("./tmp", 0755)
		if err != nil {
			log.Println("Error creating tmp directory:", err)
		}
	}
}

// Add this function to create the conversations directory structure
func ensureConversationsDirectoryExists() {
	// Create conversations directory if it doesn't exist
	if _, err := os.Stat("./conversations"); os.IsNotExist(err) {
		err := os.Mkdir("./conversations", 0755)
		if err != nil {
			log.Println("Error creating conversations directory:", err)
		}
	}
}

// Add this function to save a conversation to the conversations folder
func saveConversationToHistory(modelName string, conversation []string) error {
	// Ensure conversations directory exists
	ensureConversationsDirectoryExists()

	// Create model directory if it doesn't exist
	modelDir := filepath.Join("./conversations", modelName)
	if _, err := os.Stat(modelDir); os.IsNotExist(err) {
		err := os.Mkdir(modelDir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create model directory: %v", err)
		}
	}

	// Create a timestamped file for this conversation
	timestamp := time.Now().Format("20060102_150405")
	fileName := fmt.Sprintf("%s_%s.txt", modelName, timestamp)
	filePath := filepath.Join(modelDir, fileName)

	// Join conversation with double newlines
	content := strings.Join(conversation, "\n\n")

	// Write to file
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to write conversation file: %v", err)
	}

	return nil
}

func NewInputOutput(names []string, parent fyne.Window, settings *Settings) *InputOutput {
	// Ensure tmp directory exists
	ensureTmpDirectoryExists()

	// Ensure conversations directory exists
	ensureConversationsDirectoryExists()

	io := &InputOutput{
		OutputLabel:  widget.NewLabel(""),
		InputEntry:   widget.NewEntry(),
		ParentWindow: parent,
		Conversation: []string{},
		animating:    false,
		Settings:     settings,
	}

	// Make the output label text selectable
	io.OutputLabel.TextStyle = fyne.TextStyle{Monospace: true}
	io.OutputLabel.Wrapping = fyne.TextWrapWord

	// Set placeholder text for input
	io.InputEntry.SetPlaceHolder("Type your message here... (Press Enter to send)")

	// Create clear button
	io.ClearButton = widget.NewButton("Clear Conversation", func() {
		// Save the current conversation to history before clearing
		if len(io.Conversation) > 0 && io.ModelSelect.Selected != "" {
			err := saveConversationToHistory(io.ModelSelect.Selected, io.Conversation)
			if err != nil {
				dialog.ShowError(fmt.Errorf("Failed to save conversation history: %v", err), parent)
			}
		}

		// Get the file path for the current model
		filePath := fmt.Sprintf("./tmp/%s.txt", io.ModelSelect.Selected)

		// Delete the file content by creating an empty file
		err := os.WriteFile(filePath, []byte(""), 0644)
		if err != nil {
			dialog.ShowError(fmt.Errorf("Failed to clear chat history: %v", err), io.ParentWindow)
			return
		}

		// Clear the current conversation
		io.Conversation = []string{}
		io.OutputLabel.SetText(fmt.Sprintf("Welcome to NeuraTalk! You are now chatting with %s.\n\nType your message below to begin.", io.ModelSelect.Selected))
		io.OutputLabel.Refresh()

		// Create a new chat instance with the same model
		newIO := NewInputOutput([]string{io.ModelSelect.Selected}, io.ParentWindow, io.Settings)
		newIO.ModelSelect.SetSelected(io.ModelSelect.Selected)

		// Update the last chat in the chat manager
		if manager, ok := io.ParentWindow.(interface{ SetLastChat(*InputOutput) }); ok {
			manager.SetLastChat(newIO)
		}
	})

	modelSelect := widget.NewSelect(names, func(selected string) {
		io.SelectedModel = selected

		filePath := fmt.Sprintf("./tmp/%s.txt", selected)

		_, err := os.Stat(filePath)
		fileExists := !os.IsNotExist(err)

		if !fileExists {
			file, err := os.Create(filePath)
			if err != nil {
				msg := fmt.Sprintf("Failed to create file for model %s: %v", selected, err)
				dialog.ShowError(fmt.Errorf(msg), parent)
				log.Println(msg)
				return
			}
			file.Close()
			log.Printf("Created new file for model: %s", selected)
		}

		if fileExists {
			content, err := os.ReadFile(filePath)
			if err == nil && len(content) > 0 {
				io.Conversation = []string{}
				conversations := strings.SplitSeq(string(content), "\n\n")
				for conv := range conversations {
					if strings.TrimSpace(conv) != "" {
						io.Conversation = append(io.Conversation, conv)
					}
				}
				fullResponse := strings.Join(io.Conversation, "\n\n")
				io.OutputLabel.SetText(fullResponse)
				io.OutputLabel.Refresh()
			}
		}

		// Show welcome message for new conversations
		if len(io.Conversation) == 0 {
			io.OutputLabel.SetText(fmt.Sprintf("Welcome to NeuraTalk! You are now chatting with %s.\n\nType your message below to begin.", selected))
			io.OutputLabel.Refresh()
		}
	})

	io.ModelSelect = modelSelect

	// Add keyboard shortcuts
	io.InputEntry.OnSubmitted = func(text string) {
		if strings.TrimSpace(text) != "" {
			io.GenerateResponse()
			io.InputEntry.SetText("")
		}
	}

	return io
}

func (io *InputOutput) GetInput() string {
	return io.InputEntry.Text
}

// Modified SetOutput to animate only the new response
func (io *InputOutput) SetOutput(response string) {
	// Store current scroll position
	var scrollPos fyne.Position
	if io.ScrollContainer != nil {
		scrollPos = io.ScrollContainer.Offset
	}

	// Add the new response to conversation array
	io.Conversation = append(io.Conversation, response)

	// Start animation for the new response
	io.animateNewResponseOnly(response, scrollPos)
}

// New method to animate only the most recently added response
func (io *InputOutput) animateNewResponseOnly(newResponse string, origScrollPos fyne.Position) {
	// If already animating, stop current animation
	if io.animating && io.animationTicker != nil {
		io.animationTicker.Stop()
	}

	// Setup animation
	io.animating = true

	// Calculate where the "AI:" part starts
	aiIndex := strings.Index(newResponse, "AI: ")
	if aiIndex == -1 {
		// Fallback if format is different
		aiIndex = 0
	} else {
		aiIndex += 4 // Move past "AI: "
	}

	// Get the AI part of the response
	aiResponse := ""
	if aiIndex < len(newResponse) {
		aiResponse = newResponse[aiIndex:]
	}

	// Split into prefix (the part before AI's text) and the AI text to animate
	prefix := newResponse[:aiIndex]

	// Show everything except the AI response immediately
	fullPreviousContent := ""
	if len(io.Conversation) > 1 {
		// Join all previous conversations
		previousConvs := io.Conversation[:len(io.Conversation)-1]
		fullPreviousContent = strings.Join(previousConvs, "\n\n")

		// Add a separator
		if fullPreviousContent != "" {
			fullPreviousContent += "\n\n"
		}
	}

	// Initial display (everything except AI response)
	initialContent := fullPreviousContent + prefix
	io.OutputLabel.SetText(initialContent)
	io.OutputLabel.Refresh()

	// Respect user's scroll position
	if io.ScrollContainer != nil {
		// Only auto-scroll if user was already at bottom
		wasAtBottom := (io.ScrollContainer.Offset.Y >= io.ScrollContainer.Content.Size().Height-io.ScrollContainer.Size().Height-10)

		if !wasAtBottom && origScrollPos.Y > 0 {
			// Restore previous position
			io.ScrollContainer.Offset = origScrollPos
		} else {
			// Auto-scroll
			io.ScrollContainer.ScrollToBottom()
		}
		io.ScrollContainer.Refresh()
	}

	// Create animation ticker
	io.animationTicker = time.NewTicker(20 * time.Millisecond)

	// Animation variables
	batchSize := 3
	aiCharIndex := 0

	// Start animation in goroutine
	go func() {
		defer func() {
			io.animating = false
			if io.animationTicker != nil {
				io.animationTicker.Stop()
				io.animationTicker = nil
			}
		}()

		for range io.animationTicker.C {
			if !io.animating {
				break
			}

			// Calculate next batch
			endIdx := aiCharIndex + batchSize
			if endIdx > len(aiResponse) {
				endIdx = len(aiResponse)
			}

			// Build current display text
			currentAIText := aiResponse[:endIdx]
			displayText := initialContent + currentAIText

			// Update display
			io.OutputLabel.SetText(displayText)
			io.OutputLabel.Refresh()

			// Smart scrolling - only auto-scroll if user is already at bottom
			if io.ScrollContainer != nil {
				position := io.ScrollContainer.Offset
				contentHeight := io.ScrollContainer.Content.Size().Height
				visibleHeight := io.ScrollContainer.Size().Height

				// If close to bottom, scroll to keep up
				if position.Y >= contentHeight-visibleHeight-50 {
					io.ScrollContainer.ScrollToBottom()
				}
			}

			// Update animation progress
			aiCharIndex = endIdx

			// Exit when done
			if aiCharIndex >= len(aiResponse) {
				break
			}
		}

		// Ensure final state is displayed
		finalContent := initialContent + aiResponse
		io.OutputLabel.SetText(finalContent)
		io.OutputLabel.Refresh()
	}()
}

// Improved method to stop animation
func (io *InputOutput) stopAnimation() {
	io.animating = false
	if io.animationTicker != nil {
		io.animationTicker.Stop()
		io.animationTicker = nil
	}
}

func (io *InputOutput) GenerateResponse() {
	modelName := io.ModelSelect.Selected
	if modelName == "" {
		dialog.ShowInformation("Model Required", "Please select a model from the dropdown menu above to begin chatting.", io.ParentWindow)
		return
	}

	userPrompt := io.GetInput()
	if strings.TrimSpace(userPrompt) == "" {
		return
	}

	filePath := fmt.Sprintf("./tmp/%s.txt", modelName)

	// Disable input during generation
	io.InputEntry.Disable()
	io.ClearButton.Disable()

	// Show "thinking" indicator with better formatting
	originalConversation := make([]string, len(io.Conversation))
	copy(originalConversation, io.Conversation)

	thinkingMessage := strings.Join(io.Conversation, "\n\n") + "\n\nYou: " + userPrompt + "\n\nAI: Thinking..."
	io.OutputLabel.SetText(thinkingMessage)
	io.OutputLabel.Refresh()

	// Process in background
	go func() {
		ctx := context.Background()

		// Create Ollama instance with settings
		llm, err := ollama.New(
			ollama.WithModel(modelName),
		)
		if err != nil {
			io.OutputLabel.SetText(strings.Join(io.Conversation, "\n\n"))
			dialog.ShowError(fmt.Errorf("Failed to connect to model: %v", err), io.ParentWindow)
			io.InputEntry.Enable()
			io.ClearButton.Enable()
			return
		}

		// Apply settings to the context
		ctx = context.WithValue(ctx, "temperature", io.Settings.GetTemperature())
		ctx = context.WithValue(ctx, "top_p", io.Settings.GetTopP())
		ctx = context.WithValue(ctx, "top_k", io.Settings.GetTopK())
		ctx = context.WithValue(ctx, "num_ctx", io.Settings.GetContextLength())
		ctx = context.WithValue(ctx, "num_predict", io.Settings.GetMaxTokens())

		fullPrompt := strings.Join(io.Conversation, "\n\n")
		if fullPrompt != "" {
			fullPrompt += "\n\n"
		}
		fullPrompt += userPrompt

		response, err := llms.GenerateFromSinglePrompt(ctx, llm, fullPrompt)
		if err != nil {
			io.Conversation = originalConversation
			io.OutputLabel.SetText(strings.Join(io.Conversation, "\n\n"))
			dialog.ShowError(fmt.Errorf("Failed to generate response: %v", err), io.ParentWindow)
			io.InputEntry.Enable()
			io.ClearButton.Enable()
			return
		}

		formattedEntry := "You: " + userPrompt + "\n\nAI: " + response
		io.SetOutput(formattedEntry)
		io.InputEntry.Enable()
		io.ClearButton.Enable()

		// Save the conversation to the file
		file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			dialog.ShowError(fmt.Errorf("Failed to save conversation: %v", err), io.ParentWindow)
			return
		}
		defer file.Close()

		_, err = file.WriteString(formattedEntry + "\n\n")
		if err != nil {
			dialog.ShowError(fmt.Errorf("Failed to write conversation: %v", err), io.ParentWindow)
			return
		}

		// Save to conversations history
		err = saveConversationToHistory(modelName, io.Conversation)
		if err != nil {
			dialog.ShowError(fmt.Errorf("Failed to save to conversation history: %v", err), io.ParentWindow)
		}
	}()
}

func (io *InputOutput) GetContainer() *fyne.Container {
	io.ScrollContainer = container.NewVScroll(io.OutputLabel)
	io.ScrollContainer.SetMinSize(fyne.NewSize(400, 300))
	io.OutputLabel.Wrapping = fyne.TextWrapWord

	// Create a container for the model selection and clear button
	topBar := container.NewHBox(
		widget.NewLabel("Model:"),
		io.ModelSelect,
		widget.NewButton("Clear Chat", func() {
			// Get the file path for the current model
			filePath := fmt.Sprintf("./tmp/%s.txt", io.ModelSelect.Selected)

			// Delete the file content by creating an empty file
			err := os.WriteFile(filePath, []byte(""), 0644)
			if err != nil {
				dialog.ShowError(fmt.Errorf("Failed to clear chat history: %v", err), io.ParentWindow)
				return
			}

			// Clear the current conversation
			io.Conversation = []string{}
			io.OutputLabel.SetText(fmt.Sprintf("Welcome to NeuraTalk! You are now chatting with %s.\n\nType your message below to begin.", io.ModelSelect.Selected))
			io.OutputLabel.Refresh()

			// Create a new chat instance with the same model
			newIO := NewInputOutput([]string{io.ModelSelect.Selected}, io.ParentWindow, io.Settings)
			newIO.ModelSelect.SetSelected(io.ModelSelect.Selected)

			// Update the last chat in the chat manager
			if manager, ok := io.ParentWindow.(interface{ SetLastChat(*InputOutput) }); ok {
				manager.SetLastChat(newIO)
			}
		}),
	)

	return container.NewBorder(
		topBar,             // top
		io.InputEntry,      // bottom
		nil,                // left
		nil,                // right
		io.ScrollContainer, // center
	)
}

// Add a method to manually control animation speed
func (io *InputOutput) SetAnimationSpeed(millisPerChar int) {
	if io.animating && io.animationTicker != nil {
		io.animationTicker.Stop()
		io.animationTicker = time.NewTicker(time.Duration(millisPerChar) * time.Millisecond)
	}
}

// Add a method to skip animation
func (io *InputOutput) SkipAnimation() {
	if io.animating {
		io.stopAnimation()
	}
}

func findOllamaBinary() (string, error) {
	// Check if ollama is in PATH
	ollamaPath, err := exec.LookPath("ollama")
	if err == nil {
		return ollamaPath, nil
	}

	// Common installation paths based on OS
	var possiblePaths []string
	switch runtime.GOOS {
	case "darwin":
		possiblePaths = []string{
			"/usr/local/bin/ollama",
			"/opt/homebrew/bin/ollama",
			filepath.Join(os.Getenv("HOME"), "go/bin/ollama"),
		}
	case "windows":
		possiblePaths = []string{
			"C:\\Program Files\\Ollama\\ollama.exe",
			"C:\\Program Files (x86)\\Ollama\\ollama.exe",
			filepath.Join(os.Getenv("LOCALAPPDATA"), "ollama\\ollama.exe"),
		}
	case "linux":
		possiblePaths = []string{
			"/usr/bin/ollama",
			"/usr/local/bin/ollama",
			filepath.Join(os.Getenv("HOME"), "go/bin/ollama"),
		}
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("ollama not found in PATH or common installation locations")
}

func getOllamaModels() ([]string, error) {
	ollamaPath, err := findOllamaBinary()
	if err != nil {
		return nil, fmt.Errorf("failed to find ollama: %v", err)
	}

	// Check if ollama service is running
	checkCmd := exec.Command(ollamaPath, "list")
	var checkOut bytes.Buffer
	checkCmd.Stdout = &checkOut
	if err := checkCmd.Run(); err != nil {
		return nil, fmt.Errorf("ollama service is not running: %v", err)
	}

	// Get list of models
	cmd := exec.Command(ollamaPath, "list")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to list models: %v", err)
	}

	output := out.String()
	var names []string
	lines := strings.Split(output, "\n")
	startParsing := false
	for _, line := range lines {
		if strings.HasPrefix(line, "NAME") {
			startParsing = true
			continue
		}
		if startParsing {
			columns := strings.Fields(line)
			if len(columns) > 0 {
				names = append(names, columns[0])
			}
		}
	}

	if len(names) == 0 {
		return nil, fmt.Errorf("no models found. Please install a model using 'ollama pull <model-name>'")
	}

	return names, nil
}

// GetAvailableModels returns a list of available Ollama models
func GetAvailableModels() ([]string, error) {
	return getOllamaModels()
}
