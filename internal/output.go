package internal

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// CreateAIResponseArea creates a new AI response output area
func CreateAIResponseArea() *fyne.Container {
	responseLabel := widget.NewLabel("AI Response will appear here")
	responseContainer := container.NewVBox(responseLabel)
	return responseContainer
}
