package internal

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
)

type UI struct{}

func (u *UI) MakeUI(tabs *container.AppTabs) *fyne.Container {
	grid := container.New(layout.NewGridLayout(0), tabs)

	return grid
}
