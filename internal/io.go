package internal

import (
	"context"
	"fmt"
	"strings"

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
	io.OutputLabel.SetText(fullResponse)
	if io.OutputLabel.Text != "" {
		io.OutputLabel.Wrapping = fyne.TextWrapWord
		io.OutputLabel.Resize(io.OutputLabel.MinSize())
		io.ScrollContainer.ScrollToBottom()
	}
}

func (io *InputOutput) GenerateResponse() {
	modelName := io.ModelSelect.Selected
	if modelName == "" {
		dialog.ShowInformation("Error", "Please select a model first.", io.ParentWindow)
		return
	}

	ctx := context.Background()
	llm, err := ollama.New(ollama.WithModel(modelName))
	if err != nil {
		dialog.ShowError(err, io.ParentWindow)
		return
	}

	prompt := io.GetInput()
	fullPrompt := strings.Join(io.Conversation, "\n\n") + "\n\n" + prompt

	response, err := llms.GenerateFromSinglePrompt(ctx, llm, fullPrompt)
	if err != nil {
		dialog.ShowError(err, io.ParentWindow)
		return
	}

	fmt.Println("Response:", response)
	io.SetOutput(prompt + "\n\n" + response) // Add two-line separation after each response
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
