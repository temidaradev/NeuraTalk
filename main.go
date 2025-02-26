package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"github.com/temidaradev/NeuraTalk/internal"
)

func main() {
	a := app.New()
	w := a.NewWindow("NeuraTalk")
	w.Resize(fyne.NewSize(800, 600))

	io := internal.NewInputOutput()

	var apptabs internal.AppTabs = &internal.Sidebar{}
	tabs := apptabs.Sidebar(io.GetContainer())

	var drawable internal.Containers = &internal.UI{}
	ui := drawable.MakeUI(tabs)

	w.SetContent(ui)
	w.ShowAndRun()
}
