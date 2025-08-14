//go:build js && wasm

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"syscall/js"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type WebChatManager struct {
	Instances []*WebInputOutput
	Current   int
	Window    fyne.Window
	Sidebar   *WebSidebar
	Settings  *WebSettings
	LastChat  *WebInputOutput
	serverURL string
}

type WebInputOutput struct {
	InputEntry      *widget.Entry
	OutputLabel     *widget.Label
	ModelSelect     *widget.Select
	SelectedModel   string
	ParentWindow    fyne.Window
	ScrollContainer *container.Scroll
	Conversation    []string
	ClearButton     *widget.Button
	ServerURL       string
	Settings        *WebSettings
	animating       bool
	animationTicker *time.Ticker
}

type WebSidebar struct {
	NewChatButton  *widget.Button
	LastChatButton *widget.Button
	OptionsButton  *widget.Button
	HomeButton     *widget.Button
	TabContainer   *container.DocTabs
	MainContent    *fyne.Container
}

// SettingsData represents the structure of saved settings (stored in browser localStorage)
type SettingsData struct {
	Theme          string  `json:"theme"`
	FontSize       string  `json:"fontSize"`
	AutoScroll     bool    `json:"autoScroll"`
	AnimationSpeed float64 `json:"animationSpeed"`
	Model          string  `json:"model"`
	Temperature    float64 `json:"temperature"`
	MaxTokens      float64 `json:"maxTokens"`
	TopP           float64 `json:"topP"`
	TopK           float64 `json:"topK"`
	ContextLength  float64 `json:"contextLength"`
}

type WebSettings struct {
	Window         fyne.Window
	App            fyne.App
	ThemeSelect    *widget.Select
	FontSizeSelect *widget.Select
	AutoScroll     *widget.Check
	AnimationSpeed *widget.Slider
	// LLM Settings
	TemperatureSlider   *widget.Slider
	MaxTokensSlider     *widget.Slider
	TopPSlider          *widget.Slider
	TopKSlider          *widget.Slider
	ContextLengthSlider *widget.Slider
	ModelSelect         *widget.Select
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

func NewWebChatManager(w fyne.Window, settings *WebSettings) *WebChatManager {
	serverURL := "http://localhost:8080" // Default server URL

	manager := &WebChatManager{
		Window:    w,
		Settings:  settings,
		serverURL: serverURL,
	}

	// Get available models from server
	models, err := manager.getAvailableModelsFromServer()
	if err != nil {
		log.Printf("Failed to get models from server: %v", err)
		dialog.ShowError(fmt.Errorf("Failed to connect to server: %v", err), w)
		models = []string{"qwen3:14b", "deepseek-r1:14b", "llama3.2:latest"} // Fallback models
	}

	log.Printf("Available models: %v", models)

	// Create first chat instance
	io := NewWebInputOutput(models, w, serverURL, settings)

	manager.Instances = []*WebInputOutput{io}
	manager.Current = 0
	manager.LastChat = io

	// Create sidebar
	manager.Sidebar = &WebSidebar{}

	// Set up new chat functionality
	newChatFunc := func() {
		newIO := NewWebInputOutput(models, w, serverURL, settings)
		manager.Instances = append(manager.Instances, newIO)
		manager.Current = len(manager.Instances) - 1
		manager.LastChat = newIO

		chatTab := container.NewTabItemWithIcon(
			fmt.Sprintf("Chat %d", len(manager.Instances)),
			theme.DocumentIcon(),
			newIO.GetContainer(),
		)
		manager.Sidebar.TabContainer.Append(chatTab)
		manager.Sidebar.TabContainer.Select(chatTab)
	}

	// Set up last chat functionality
	lastChatFunc := func() {
		if manager.LastChat != nil {
			for _, tab := range manager.Sidebar.TabContainer.Items {
				if tab.Content == manager.LastChat.GetContainer() {
					manager.Sidebar.TabContainer.Select(tab)
					return
				}
			}

			chatTab := container.NewTabItemWithIcon(
				"Last Chat",
				theme.DocumentIcon(),
				manager.LastChat.GetContainer(),
			)
			manager.Sidebar.TabContainer.Append(chatTab)
			manager.Sidebar.TabContainer.Select(chatTab)
		}
	}

	manager.Sidebar.NewChatButton = widget.NewButtonWithIcon("New Chat", theme.ContentAddIcon(), newChatFunc)
	manager.Sidebar.LastChatButton = widget.NewButtonWithIcon("Last Chat", theme.DocumentIcon(), lastChatFunc)

	return manager
}

// SetLastChat updates the last chat instance
func (m *WebChatManager) SetLastChat(io *WebInputOutput) {
	m.LastChat = io
	m.Instances = append(m.Instances, io)
	m.Current = len(m.Instances) - 1

	// Add new chat tab and switch to it
	chatTab := container.NewTabItemWithIcon(
		"Last Chat",
		theme.DocumentIcon(),
		io.GetContainer(),
	)
	m.Sidebar.TabContainer.Append(chatTab)
	m.Sidebar.TabContainer.Select(chatTab)
}

func (m *WebChatManager) getAvailableModelsFromServer() ([]string, error) {
	// Use JavaScript fetch API to get models from server
	promise := js.Global().Call("fetch", m.serverURL+"/api/models")

	// Create a channel to wait for the promise
	resultChan := make(chan []string, 1)
	errorChan := make(chan error, 1)

	// Handle the promise
	successHandler := js.FuncOf(func(this js.Value, p []js.Value) interface{} {
		response := p[0]
		if !response.Get("ok").Bool() {
			errorChan <- fmt.Errorf("server returned error: %d", response.Get("status").Int())
			return nil
		}

		jsonPromise := response.Call("json")
		jsonHandler := js.FuncOf(func(this js.Value, p []js.Value) interface{} {
			defer func() {
				if r := recover(); r != nil {
					errorChan <- fmt.Errorf("error parsing JSON: %v", r)
				}
			}()

			data := p[0]
			modelsJS := data.Get("models")

			var models []string
			if modelsJS.Type() != js.TypeUndefined && !modelsJS.IsNull() {
				length := modelsJS.Get("length").Int()
				for i := 0; i < length; i++ {
					model := modelsJS.Index(i).String()
					if model != "" {
						models = append(models, model)
					}
				}
			}

			if len(models) == 0 {
				models = []string{"No models available"}
			}

			resultChan <- models
			return nil
		})

		jsonErrorHandler := js.FuncOf(func(this js.Value, p []js.Value) interface{} {
			errorChan <- fmt.Errorf("failed to parse JSON response")
			return nil
		})

		jsonPromise.Call("then", jsonHandler)
		jsonPromise.Call("catch", jsonErrorHandler)
		return nil
	})

	// Handle errors
	errorHandler := js.FuncOf(func(this js.Value, p []js.Value) interface{} {
		errorMsg := "fetch error"
		if len(p) > 0 && !p[0].IsNull() {
			if msg := p[0].Get("message"); !msg.IsUndefined() {
				errorMsg = msg.String()
			}
		}
		errorChan <- fmt.Errorf("fetch error: %s", errorMsg)
		return nil
	})

	promise.Call("then", successHandler)
	promise.Call("catch", errorHandler)

	// Wait for result or error with timeout
	select {
	case models := <-resultChan:
		return models, nil
	case err := <-errorChan:
		return []string{"Connection error"}, err
	case <-time.After(10 * time.Second):
		return []string{"Timeout error"}, fmt.Errorf("request timed out")
	}
}

func NewWebInputOutput(names []string, parent fyne.Window, serverURL string, settings *WebSettings) *WebInputOutput {
	io := &WebInputOutput{
		OutputLabel:  widget.NewLabel(""),
		InputEntry:   widget.NewEntry(),
		ParentWindow: parent,
		Conversation: []string{},
		ServerURL:    serverURL,
		Settings:     settings,
		animating:    false,
	}

	io.OutputLabel.TextStyle = fyne.TextStyle{Monospace: true}
	io.OutputLabel.Wrapping = fyne.TextWrapWord
	io.InputEntry.SetPlaceHolder("Type your message here... (Press Enter to send)")

	// Create clear button with conversation history saving (simulated for web)
	io.ClearButton = widget.NewButton("Clear Conversation", func() {
		// Save the current conversation to localStorage before clearing (simulated)
		if len(io.Conversation) > 0 && io.ModelSelect.Selected != "" {
			// In a real implementation, you'd save to localStorage here
			log.Printf("Would save conversation history for model: %s", io.ModelSelect.Selected)
		}

		// Clear the current conversation
		io.Conversation = []string{}
		io.OutputLabel.SetText(fmt.Sprintf("Welcome to NeuraTalk Web! You are now chatting with %s.\n\nType your message below to begin.", io.ModelSelect.Selected))
		io.OutputLabel.Refresh()
	})

	modelSelect := widget.NewSelect(names, func(selected string) {
		io.SelectedModel = selected

		// For web version, we simulate loading conversation history
		// In desktop version, this loads from files
		if len(io.Conversation) == 0 {
			io.OutputLabel.SetText(fmt.Sprintf("Welcome to NeuraTalk Web! You are now chatting with %s.\n\nType your message below to begin.", selected))
			io.OutputLabel.Refresh()
		}
	})

	io.ModelSelect = modelSelect

	// Auto-select the first model if available
	if len(names) > 0 && names[0] != "No models available" {
		modelSelect.SetSelected(names[0])
		io.SelectedModel = names[0]
	}

	// Add keyboard shortcuts
	io.InputEntry.OnSubmitted = func(text string) {
		if strings.TrimSpace(text) != "" {
			io.GenerateResponse()
			io.InputEntry.SetText("")
		}
	}

	return io
}

func (io *WebInputOutput) GetInput() string {
	return io.InputEntry.Text
}

func (io *WebInputOutput) SetOutput(response string) {
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

// Method to animate only the most recently added response
func (io *WebInputOutput) animateNewResponseOnly(newResponse string, origScrollPos fyne.Position) {
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

	// Get animation speed from settings
	animSpeed := 20 * time.Millisecond
	if io.Settings != nil {
		animSpeed = time.Duration(io.Settings.GetAnimationSpeed()) * time.Millisecond
	}

	// Create animation ticker
	io.animationTicker = time.NewTicker(animSpeed)

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

func (io *WebInputOutput) GenerateResponse() {
	modelName := io.ModelSelect.Selected
	if modelName == "" {
		dialog.ShowInformation("Model Required", "Please select a model from the dropdown menu above to begin chatting.", io.ParentWindow)
		return
	}

	userPrompt := io.GetInput()
	if strings.TrimSpace(userPrompt) == "" {
		return
	}

	// Disable input during generation
	io.InputEntry.Disable()
	io.ClearButton.Disable()

	// Show "thinking" indicator with better formatting
	originalConversation := make([]string, len(io.Conversation))
	copy(originalConversation, io.Conversation)

	thinkingMessage := strings.Join(io.Conversation, "\n\n") + "\n\nYou: " + userPrompt + "\n\nAI: Thinking..."
	io.OutputLabel.SetText(thinkingMessage)
	io.OutputLabel.Refresh()

	// Prepare request with settings
	request := ChatRequest{
		Model:   modelName,
		Message: userPrompt,
		History: io.Conversation,
	}

	// Add settings to request if available
	if io.Settings != nil {
		// In a real implementation, you would add temperature, maxTokens, etc. to the request
		log.Printf("Using settings - Temp: %.2f, MaxTokens: %.0f", io.Settings.GetTemperature(), io.Settings.GetMaxTokens())
	}

	requestJSON, _ := json.Marshal(request)

	// Create fetch options
	options := js.Global().Get("Object").New()
	options.Set("method", "POST")
	options.Set("headers", js.Global().Get("Object").New())
	options.Get("headers").Set("Content-Type", "application/json")
	options.Set("body", string(requestJSON))

	// Make request to server
	promise := js.Global().Call("fetch", io.ServerURL+"/api/chat", options)

	successHandler := js.FuncOf(func(this js.Value, p []js.Value) interface{} {
		response := p[0]
		if !response.Get("ok").Bool() {
			io.handleError(fmt.Sprintf("Server returned error: %d", response.Get("status").Int()))
			return nil
		}

		jsonPromise := response.Call("json")
		jsonHandler := js.FuncOf(func(this js.Value, p []js.Value) interface{} {
			defer func() {
				if r := recover(); r != nil {
					io.handleError(fmt.Sprintf("Error parsing response: %v", r))
				}
			}()

			data := p[0]
			responseText := ""
			errorText := ""

			if respField := data.Get("response"); !respField.IsUndefined() && !respField.IsNull() {
				responseText = respField.String()
			}

			if errField := data.Get("error"); !errField.IsUndefined() && !errField.IsNull() {
				errorText = errField.String()
			}

			if errorText != "" {
				io.handleError(errorText)
				return nil
			}

			if responseText == "" {
				io.handleError("Empty response from server")
				return nil
			}

			// Format the response like the desktop version
			formattedEntry := "You: " + userPrompt + "\n\nAI: " + responseText
			io.SetOutput(formattedEntry)
			io.InputEntry.Enable()
			io.ClearButton.Enable()
			return nil
		})

		jsonErrorHandler := js.FuncOf(func(this js.Value, p []js.Value) interface{} {
			io.handleError("Failed to parse server response")
			return nil
		})

		jsonPromise.Call("then", jsonHandler)
		jsonPromise.Call("catch", jsonErrorHandler)
		return nil
	})

	errorHandler := js.FuncOf(func(this js.Value, p []js.Value) interface{} {
		io.Conversation = originalConversation
		errorMsg := "Network error occurred"
		if len(p) > 0 && !p[0].IsNull() {
			if msg := p[0].Get("message"); !msg.IsUndefined() {
				errorMsg = fmt.Sprintf("Network error: %s", msg.String())
			}
		}
		io.handleError(errorMsg)
		return nil
	})

	promise.Call("then", successHandler)
	promise.Call("catch", errorHandler)
}

func (io *WebInputOutput) handleError(errorMsg string) {
	io.OutputLabel.SetText(strings.Join(io.Conversation, "\n\n"))
	io.OutputLabel.Refresh()
	dialog.ShowError(fmt.Errorf("%s", errorMsg), io.ParentWindow)
	io.InputEntry.Enable()
	io.ClearButton.Enable()
}

func (io *WebInputOutput) GetContainer() *fyne.Container {
	io.ScrollContainer = container.NewVScroll(io.OutputLabel)
	io.ScrollContainer.SetMinSize(fyne.NewSize(400, 300))
	io.OutputLabel.Wrapping = fyne.TextWrapWord

	// Create a container for the model selection and clear button
	topBar := container.NewHBox(
		widget.NewLabel("Model:"),
		io.ModelSelect,
		widget.NewButton("Clear Chat", func() {
			// Save conversation to localStorage before clearing (simulated)
			if len(io.Conversation) > 0 && io.ModelSelect.Selected != "" {
				log.Printf("Saving conversation history for model: %s", io.ModelSelect.Selected)
			}

			// Clear the current conversation
			io.Conversation = []string{}
			io.OutputLabel.SetText(fmt.Sprintf("Welcome to NeuraTalk Web! You are now chatting with %s.\n\nType your message below to begin.", io.ModelSelect.Selected))
			io.OutputLabel.Refresh()
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

func (s *WebSidebar) Sidebar(cont *fyne.Container, settings *fyne.Container) *container.Split {
	// Create main content area with closeable tabs
	s.TabContainer = container.NewDocTabs()
	s.MainContent = cont

	// Create welcome content
	welcomeTitle := widget.NewLabelWithStyle("Welcome to NeuraTalk Web", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	welcomeTitle.TextStyle.Bold = true

	welcomeText := widget.NewTextGridFromString(
		"Your AI companion for seamless conversations.\n\n" +
			"Start a new chat to begin your journey!")

	// Create Last Chat button for the home menu
	welcomeLastChatBtn := widget.NewButtonWithIcon("Last Chat", theme.DocumentIcon(), func() {
		if s.LastChatButton != nil && s.LastChatButton.OnTapped != nil {
			s.LastChatButton.OnTapped()
		}
	})

	welcomeNewChatBtn := widget.NewButtonWithIcon("New Chat", theme.ContentAddIcon(), func() {
		if s.NewChatButton != nil && s.NewChatButton.OnTapped != nil {
			s.NewChatButton.OnTapped()
		}
	})

	welcomeContent := container.NewVBox(
		layout.NewSpacer(),
		welcomeTitle,
		welcomeText,
		layout.NewSpacer(),
		container.NewHBox(
			layout.NewSpacer(),
			welcomeLastChatBtn,
			layout.NewSpacer(),
		),
		container.NewHBox(
			layout.NewSpacer(),
			welcomeNewChatBtn,
			layout.NewSpacer(),
		),
		layout.NewSpacer(),
	)

	// Create home button
	s.HomeButton = widget.NewButtonWithIcon("Home", theme.HomeIcon(), func() {
		// Check if home tab already exists
		for _, tab := range s.TabContainer.Items {
			if tab.Text == "Home" {
				s.TabContainer.Select(tab)
				return
			}
		}
		// Create new home tab
		homeTab := container.NewTabItemWithIcon("Home", theme.HomeIcon(), welcomeContent)
		s.TabContainer.Append(homeTab)
		s.TabContainer.Select(homeTab)
	})

	// Create options button
	s.OptionsButton = widget.NewButtonWithIcon("Options", theme.SettingsIcon(), func() {
		// Check if options tab already exists
		for _, tab := range s.TabContainer.Items {
			if tab.Text == "Options" {
				s.TabContainer.Select(tab)
				return
			}
		}
		// Create new options tab
		optionsTab := container.NewTabItemWithIcon("Options", theme.SettingsIcon(), settings)
		s.TabContainer.Append(optionsTab)
		s.TabContainer.Select(optionsTab)
	})

	// Create sidebar content
	topContent := container.NewVBox(
		s.HomeButton,
		widget.NewSeparator(),
		s.NewChatButton,
		widget.NewSeparator(),
		s.OptionsButton,
	)

	// Add padding around the buttons
	paddedContent := container.NewPadded(topContent)

	// Create a split container with resizable sidebar
	split := container.NewHSplit(
		paddedContent,
		s.TabContainer,
	)
	split.SetOffset(0.2) // Set initial sidebar width to 20% of window width

	// Show home tab by default
	s.HomeButton.OnTapped()

	return split
}

func NewWebSettings(w fyne.Window, a fyne.App) *WebSettings {
	s := &WebSettings{
		Window: w,
		App:    a,
	}

	// Initialize UI elements first
	s.initializeUIElements()

	// Then load saved settings from localStorage
	s.loadSettings()

	return s
}

func (s *WebSettings) initializeUIElements() {
	// Theme selection
	s.ThemeSelect = widget.NewSelect([]string{"Light", "Dark"}, func(selected string) {
		if selected == "Dark" {
			s.App.Settings().SetTheme(theme.DarkTheme())
		} else {
			s.App.Settings().SetTheme(theme.LightTheme())
		}
		s.saveSettings()
	})

	// Font size selection
	s.FontSizeSelect = widget.NewSelect([]string{"Small", "Medium", "Large"}, func(selected string) {
		s.saveSettings()
	})

	// Auto-scroll toggle
	s.AutoScroll = widget.NewCheck("Auto-scroll to new messages", func(checked bool) {
		s.saveSettings()
	})

	// Animation speed slider
	s.AnimationSpeed = widget.NewSlider(10, 100)
	s.AnimationSpeed.OnChanged = func(value float64) {
		s.saveSettings()
	}

	// Model selection
	s.ModelSelect = widget.NewSelect([]string{}, func(selected string) {
		s.saveSettings()
	})

	// Temperature slider
	s.TemperatureSlider = widget.NewSlider(0, 2)
	s.TemperatureSlider.OnChanged = func(value float64) {
		s.saveSettings()
	}

	// Max tokens slider
	s.MaxTokensSlider = widget.NewSlider(100, 4096)
	s.MaxTokensSlider.OnChanged = func(value float64) {
		s.saveSettings()
	}

	// Top P slider
	s.TopPSlider = widget.NewSlider(0, 1)
	s.TopPSlider.OnChanged = func(value float64) {
		s.saveSettings()
	}

	// Top K slider
	s.TopKSlider = widget.NewSlider(1, 100)
	s.TopKSlider.OnChanged = func(value float64) {
		s.saveSettings()
	}

	// Context length slider
	s.ContextLengthSlider = widget.NewSlider(512, 8192)
	s.ContextLengthSlider.OnChanged = func(value float64) {
		s.saveSettings()
	}
}

func (s *WebSettings) saveSettings() {
	settings := SettingsData{
		Theme:          s.ThemeSelect.Selected,
		FontSize:       s.FontSizeSelect.Selected,
		AutoScroll:     s.AutoScroll.Checked,
		AnimationSpeed: s.AnimationSpeed.Value,
		Model:          s.ModelSelect.Selected,
		Temperature:    s.TemperatureSlider.Value,
		MaxTokens:      s.MaxTokensSlider.Value,
		TopP:           s.TopPSlider.Value,
		TopK:           s.TopKSlider.Value,
		ContextLength:  s.ContextLengthSlider.Value,
	}

	// Save to localStorage (browser storage)
	data, err := json.Marshal(settings)
	if err != nil {
		log.Printf("Failed to marshal settings: %v", err)
		return
	}

	// Use JavaScript to save to localStorage
	js.Global().Get("localStorage").Call("setItem", "neuratalk_settings", string(data))
}

func (s *WebSettings) loadSettings() {
	// Default settings
	defaultSettings := SettingsData{
		Theme:          "Light",
		FontSize:       "Medium",
		AutoScroll:     true,
		AnimationSpeed: 20,
		Model:          "",
		Temperature:    0.7,
		MaxTokens:      2048,
		TopP:           0.9,
		TopK:           40,
		ContextLength:  4096,
	}

	// Try to load saved settings from localStorage
	settingsData := js.Global().Get("localStorage").Call("getItem", "neuratalk_settings")
	if !settingsData.IsNull() {
		var settings SettingsData
		if err := json.Unmarshal([]byte(settingsData.String()), &settings); err == nil {
			defaultSettings = settings
		}
	}

	// Apply loaded settings to UI elements
	if s.ThemeSelect != nil {
		s.ThemeSelect.SetSelected(defaultSettings.Theme)
		// Apply theme immediately
		if defaultSettings.Theme == "Dark" {
			s.App.Settings().SetTheme(theme.DarkTheme())
		} else {
			s.App.Settings().SetTheme(theme.LightTheme())
		}
	}

	if s.FontSizeSelect != nil {
		s.FontSizeSelect.SetSelected(defaultSettings.FontSize)
	}

	if s.AutoScroll != nil {
		s.AutoScroll.SetChecked(defaultSettings.AutoScroll)
	}

	if s.AnimationSpeed != nil {
		s.AnimationSpeed.SetValue(defaultSettings.AnimationSpeed)
	}

	if s.ModelSelect != nil {
		s.ModelSelect.SetSelected(defaultSettings.Model)
	}

	if s.TemperatureSlider != nil {
		s.TemperatureSlider.SetValue(defaultSettings.Temperature)
	}

	if s.MaxTokensSlider != nil {
		s.MaxTokensSlider.SetValue(defaultSettings.MaxTokens)
	}

	if s.TopPSlider != nil {
		s.TopPSlider.SetValue(defaultSettings.TopP)
	}

	if s.TopKSlider != nil {
		s.TopKSlider.SetValue(defaultSettings.TopK)
	}

	if s.ContextLengthSlider != nil {
		s.ContextLengthSlider.SetValue(defaultSettings.ContextLength)
	}
}

// Settings getters
func (s *WebSettings) GetTheme() string {
	return s.ThemeSelect.Selected
}

func (s *WebSettings) GetFontSize() string {
	return s.FontSizeSelect.Selected
}

func (s *WebSettings) IsAutoScrollEnabled() bool {
	return s.AutoScroll.Checked
}

func (s *WebSettings) GetAnimationSpeed() float64 {
	return s.AnimationSpeed.Value
}

func (s *WebSettings) GetModel() string {
	return s.ModelSelect.Selected
}

func (s *WebSettings) GetTemperature() float64 {
	return s.TemperatureSlider.Value
}

func (s *WebSettings) GetMaxTokens() float64 {
	return s.MaxTokensSlider.Value
}

func (s *WebSettings) GetTopP() float64 {
	return s.TopPSlider.Value
}

func (s *WebSettings) GetTopK() float64 {
	return s.TopKSlider.Value
}

func (s *WebSettings) GetContextLength() float64 {
	return s.ContextLengthSlider.Value
}

func (s *WebSettings) GetContainer() *fyne.Container {
	// Create labels for the settings
	themeLabel := widget.NewLabel("Theme:")
	fontLabel := widget.NewLabel("Font Size:")
	llmSettingsLabel := widget.NewLabelWithStyle("LLM Settings", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	modelLabel := widget.NewLabel("Model:")

	// Create a container with the form and model config button
	content := container.NewVBox(
		widget.NewLabel("Settings"),
		widget.NewSeparator(),
		container.NewVBox(
			widget.NewLabel("Theme"),
			container.NewHBox(themeLabel, s.ThemeSelect),
			widget.NewLabel("Font Size"),
			container.NewHBox(fontLabel, s.FontSizeSelect),
			s.AutoScroll,
			widget.NewLabel("Animation Speed"),
			s.AnimationSpeed,
			widget.NewLabel("(slower) ← → (faster)"),
		),
		widget.NewSeparator(),
		llmSettingsLabel,
		widget.NewSeparator(),
		container.NewVBox(
			widget.NewLabel("Model"),
			container.NewHBox(modelLabel, s.ModelSelect),
			widget.NewLabel("Temperature"),
			s.TemperatureSlider,
			widget.NewLabel("(deterministic) ← → (creative)"),
			widget.NewLabel("Max Tokens"),
			s.MaxTokensSlider,
			widget.NewLabel("(shorter) ← → (longer)"),
			widget.NewLabel("Top P"),
			s.TopPSlider,
			widget.NewLabel("(focused) ← → (diverse)"),
			widget.NewLabel("Top K"),
			s.TopKSlider,
			widget.NewLabel("(focused) ← → (diverse)"),
			widget.NewLabel("Context Length"),
			s.ContextLengthSlider,
			widget.NewLabel("(shorter) ← → (longer)"),
		),
	)

	// Wrap the content in a scroll container
	scrollContent := container.NewVScroll(content)
	scrollContent.SetMinSize(fyne.NewSize(300, 400))

	return container.NewBorder(
		nil,           // top
		nil,           // bottom
		nil,           // left
		nil,           // right
		scrollContent, // center
	)
}

func main() {
	a := app.New()
	w := a.NewWindow("NeuraTalk Web")
	w.Resize(fyne.NewSize(1000, 700))

	// Set initial theme before creating any widgets
	if settings := a.Settings(); settings != nil {
		settings.SetTheme(theme.LightTheme())
	}

	// Create web settings
	webSettings := NewWebSettings(w, a)

	// Create chat manager
	manager := NewWebChatManager(w, webSettings)

	// Create initial UI
	split := manager.Sidebar.Sidebar(nil, webSettings.GetContainer())
	w.SetContent(split)

	// Start with the last chat if it exists
	if manager.LastChat != nil {
		chatTab := container.NewTabItemWithIcon(
			"Chat",
			theme.DocumentIcon(),
			manager.LastChat.GetContainer(),
		)
		manager.Sidebar.TabContainer.Append(chatTab)
		manager.Sidebar.TabContainer.Select(chatTab)
	}

	w.CenterOnScreen()

	// Log that we're running in WASM mode
	log.Println("NeuraTalk Web running in WASM mode")

	w.ShowAndRun()
}
