package internal

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
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
	})

	return &InputOutput{
		OutputLabel:  widget.NewLabel("AI Response will appear here"),
		InputEntry:   widget.NewEntry(),
		ModelSelect:  modelSelect,
		ParentWindow: parent,
	}
}

func (io *InputOutput) GetInput() string {
	return io.InputEntry.Text
}

func (io *InputOutput) SetOutput(response string) {
	io.OutputLabel.SetText(response)
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
