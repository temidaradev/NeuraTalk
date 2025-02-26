package internal

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type Sidebar struct{}

func (s *Sidebar) Sidebar(cont *fyne.Container) *container.AppTabs {
	tabs := container.NewAppTabs(
		container.NewTabItem("AI Response", cont),
		container.NewTabItem("Tab 1", widget.NewLabel("Hello")),
		container.NewTabItem("Tab 2", widget.NewLabel("World!")),
	)

	tabs.SetTabLocation(container.TabLocationLeading)

	return tabs
}
