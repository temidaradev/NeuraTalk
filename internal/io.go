package internal

import (
	"context"
	"fmt"
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
	// Animation control
	animating       bool
	animationTicker *time.Ticker
	currentText     string
	targetText      string
	charIndex       int
}

func NewInputOutput(names []string, parent fyne.Window) *InputOutput {
	modelSelect := widget.NewSelect(names, func(selected string) {
		dialog.ShowInformation("Model Selected", "Selected model: "+selected, parent)
		fmt.Println("Selected model:", selected)
	})

	io := &InputOutput{
		OutputLabel:  widget.NewLabel(""),
		InputEntry:   widget.NewEntry(),
		ModelSelect:  modelSelect,
		ParentWindow: parent,
		Conversation: []string{},
		animating:    false,
	}

	io.InputEntry.OnSubmitted = func(text string) {
		io.GenerateResponse()
		io.InputEntry.SetText("")
	}
	return io
}

func (io *InputOutput) GetInput() string {
	return io.InputEntry.Text
}

func (io *InputOutput) SetOutput(response string) {
	io.Conversation = append(io.Conversation, response)
	fullResponse := strings.Join(io.Conversation, "\n\n")

	// Instead of setting text directly, animate it
	io.animateText(fullResponse)
}

// New method to handle text animation
func (io *InputOutput) animateText(targetText string) {
	// If already animating, stop current animation
	io.stopAnimation()

	// Setup new animation
	io.animating = true
	io.targetText = targetText
	io.currentText = io.OutputLabel.Text
	io.charIndex = len(io.currentText)

	// Create animation ticker (adjust speed as needed)
	io.animationTicker = time.NewTicker(30 * time.Millisecond)

	// Start animation in a goroutine
	go func() {
		for range io.animationTicker.C {
			if !io.animating || io.charIndex >= len(io.targetText) {
				io.stopAnimation()
				return
			}

			// Show one more character
			io.charIndex++
			displayText := io.targetText[:io.charIndex]

			// Update UI on main thread
			fyne.CurrentApp().Driver().CanvasForObject(io.OutputLabel).Content().Refresh()

			// Use a separate goroutine to safely update the UI
			go func(text string) {
				// Use a small sleep to avoid potential race conditions
				time.Sleep(1 * time.Millisecond)
				io.OutputLabel.SetText(text)
				io.OutputLabel.Wrapping = fyne.TextWrapWord
				io.OutputLabel.Resize(io.OutputLabel.MinSize())
				io.ScrollContainer.ScrollToBottom()
				io.OutputLabel.Refresh()
			}(displayText)
		}
	}()
}

// New method to stop ongoing animation
func (io *InputOutput) stopAnimation() {
	if io.animating && io.animationTicker != nil {
		io.animationTicker.Stop()
		io.animating = false

		// Ensure text is fully displayed when stopping animation
		if io.targetText != "" {
			io.OutputLabel.SetText(io.targetText)
			io.OutputLabel.Wrapping = fyne.TextWrapWord
			io.OutputLabel.Resize(io.OutputLabel.MinSize())
			io.ScrollContainer.ScrollToBottom()
			io.OutputLabel.Refresh()
		}
	}
}

func (io *InputOutput) GenerateResponse() {
	modelName := io.ModelSelect.Selected
	if modelName == "" {
		dialog.ShowInformation("Error", "Please select a model first.", io.ParentWindow)
		return
	}

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
