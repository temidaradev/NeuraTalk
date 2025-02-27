package internal

import (
	"context"
	"fmt"

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
	}

	return io
}

func (io *InputOutput) GetInput() string {
	return io.InputEntry.Text
}

func (io *InputOutput) SetOutput(response string) {
	io.OutputLabel.SetText(response)
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
	response, err := llms.GenerateFromSinglePrompt(ctx, llm, prompt)
	if err != nil {
		dialog.ShowError(err, io.ParentWindow)
		return
	}

	io.SetOutput(response)
}

func (io *InputOutput) GetContainer() *fyne.Container {
	return container.NewBorder(
		io.ModelSelect, // top
		io.InputEntry,  // bottom
		nil,            // left
		nil,            // right
		io.OutputLabel, // center
	)
}
