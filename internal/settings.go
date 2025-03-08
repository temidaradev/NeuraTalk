package internal

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type Settings struct {
}

func NewSettings() *Settings {
	return &Settings{}
}

func (s *Settings) GetLabel() *widget.Label {
	label := widget.NewLabel("Settings")
	return label
}

func (s *Settings) GetContainer() *fyne.Container {
	return container.NewBorder(
		s.GetLabel(), // top
		nil,          // bottom
		nil,          // left
		nil,          // right
		nil,          // center
	)
}
