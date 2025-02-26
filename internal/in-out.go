package internal

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type InputOutput struct {
	InputEntry  *widget.Entry
	OutputLabel *widget.Label
}

func NewInputOutput() *InputOutput {
	return &InputOutput{
		OutputLabel: widget.NewLabel("AI Response will appear here"),
		InputEntry:  widget.NewEntry(),
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
		nil,            // top
		io.InputEntry,  // bottom
		nil,            // left
		nil,            // right
		io.OutputLabel, // center
	)
}
