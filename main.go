package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/temidaradev/AIChatGUI/internal"
)

func main() {
	a := app.New()
	w := a.NewWindow("NeuraTalk")
	w.Resize(fyne.NewSize(800, 600))

	// var drawable internal.Containers = &internal.UI{}
	// ui := drawable.MakeUI()
	var apptabs internal.AppTabs = &internal.Sidebar{}
	tabs := apptabs.Sidebar()

	w.SetContent(tabs)
	w.ShowAndRun()
}
