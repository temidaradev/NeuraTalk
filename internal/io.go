package internal

import (
	"context"
	"fmt"
	"log"
	"os"
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

func NewInputOutput(names []string, parent fyne.Window) *InputOutput {
	// Ensure tmp directory exists
	ensureTmpDirectoryExists()

	io := &InputOutput{
		OutputLabel:  widget.NewLabel(""),
		InputEntry:   widget.NewEntry(),
		ParentWindow: parent,
		Conversation: []string{},
		animating:    false,
	}

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

		dialog.ShowInformation("Model Selected", "Selected model: "+selected, parent)
		fmt.Println("Selected model:", selected)
	})

	io.ModelSelect = modelSelect

	io.InputEntry.OnSubmitted = func(text string) {
		io.GenerateResponse()
		io.InputEntry.SetText("")
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
		dialog.ShowInformation("Error", "Please select a model first.", io.ParentWindow)
		return
	}
	filePath := fmt.Sprintf("./tmp/%s.txt", modelName)

	// Disable input during generation
	io.InputEntry.Disable()

	// Show "thinking" indicator
	userPrompt := io.GetInput()
	originalConversation := make([]string, len(io.Conversation))
	copy(originalConversation, io.Conversation)

	// Add thinking indicator without changing the actual conversation
	io.OutputLabel.SetText(strings.Join(io.Conversation, "\n\n") + "\n\nYou: " + userPrompt + "\n\nAI: Thinking...")
	io.OutputLabel.Refresh()

	// Process in background
	go func() {
		ctx := context.Background()
		llm, err := ollama.New(ollama.WithModel(modelName))
		if err != nil {
			// We need to handle UI updates on the main thread
			io.OutputLabel.SetText(strings.Join(io.Conversation, "\n\n"))
			dialog.ShowError(err, io.ParentWindow)
			io.InputEntry.Enable()
			return
		}

		fullPrompt := strings.Join(io.Conversation, "\n\n")
		if fullPrompt != "" {
			fullPrompt += "\n\n"
		}
		fullPrompt += userPrompt

		response, err := llms.GenerateFromSinglePrompt(ctx, llm, fullPrompt)
		if err != nil {
			// Restore original conversation on error
			io.Conversation = originalConversation
			io.OutputLabel.SetText(strings.Join(io.Conversation, "\n\n"))
			dialog.ShowError(err, io.ParentWindow)
			io.InputEntry.Enable()
			return
		}

		// Format and add the new conversation entry
		formattedEntry := "You: " + userPrompt + "\n\nAI: " + response
		io.SetOutput(formattedEntry)
		io.InputEntry.Enable()

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

		fmt.Println(formattedEntry)
	}()
}

func (io *InputOutput) GetContainer() *fyne.Container {
	io.ScrollContainer = container.NewVScroll(io.OutputLabel)
	io.ScrollContainer.SetMinSize(fyne.NewSize(400, 300)) // Set a minimum size for the scroll container
	io.OutputLabel.Wrapping = fyne.TextWrapWord           // Ensure text wrapping

	return container.NewBorder(
		io.ModelSelect,     // top
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
