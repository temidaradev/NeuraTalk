package internal

import (
	"context"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

type InputOutput struct {
	InputEntry    *widget.Entry
	OutputLabel   *widget.Label
	ModelSelect   *widget.Select
	SelectedModel string
	ParentWindow  fyne.Window
}

func NewInputOutput(names []string, parent fyne.Window) *InputOutput {
	modelSelect := widget.NewSelect(names, func(selected string) {
		dialog.ShowInformation("Model Selected", "Selected model: "+selected, parent)
		fmt.Println("Selected model:", selected)
	})

	io := &InputOutput{
		OutputLabel:  widget.NewLabel("AI Response will appear here"),
		InputEntry:   widget.NewEntry(),
		ModelSelect:  modelSelect,
		ParentWindow: parent,
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
	io.OutputLabel.SetText(response)
	if io.OutputLabel.Text != "" {
		io.OutputLabel.Wrapping = fyne.TextWrapWord
		io.OutputLabel.Resize(io.OutputLabel.MinSize())
	}
}

func (io *InputOutput) GenerateResponse() {
	modelName := io.ModelSelect.Selected

	ctx := context.Background()
	llm, err := ollama.New(ollama.WithModel(modelName))
	if err != nil {
		dialog.ShowError(err, io.ParentWindow)
		return
	}

	prompt := io.GetInput()

	// Start sliding text animation
	stopChan := make(chan struct{})
	go io.startSlidingText("Waiting for response", stopChan)

	response, err := llms.GenerateFromSinglePrompt(ctx, llm, prompt)
	close(stopChan) // Stop the sliding text animation

	if err != nil {
		dialog.ShowError(err, io.ParentWindow)
		return
	}

	io.SetOutput(response)
}

func (io *InputOutput) startSlidingText(baseText string, stopChan chan struct{}) {
	ticker := time.NewTicker(300 * time.Millisecond)
	defer ticker.Stop()

	dots := ""
	for {
		select {
		case <-ticker.C:
			dots += "."
			if len(dots) > 3 {
				dots = ""
			}
			io.OutputLabel.SetText(baseText + dots)
		case <-stopChan:
			return
		}
	}
}

func (io *InputOutput) GetContainer() *fyne.Container {
	scrollContainer := container.NewScroll(io.OutputLabel)
	return container.NewBorder(
		io.ModelSelect,  // top
		io.InputEntry,   // bottom
		nil,             // left
		nil,             // right
		scrollContainer, // center
	)
}
