package internal

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type Selection struct {
	ModelSelect   *widget.Select
	SelectedModel string
}

func NewSelection(names []string) *Selection {
	modelSelect := widget.NewSelect(names, func(selected string) {
		dialog.ShowInformation("Model Selected", "Selected model: "+selected, nil)
	})

	return &Selection{
		ModelSelect: modelSelect,
	}
}

func (s *Selection) GetContainer() *fyne.Container {
	return container.NewVBox(
		widget.NewLabel("Model:"),
		s.ModelSelect,
	)
}
