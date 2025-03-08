package internal

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

type Sidebar struct{}

func (s *Sidebar) Sidebar(cont *fyne.Container, settings *fyne.Container) *container.AppTabs {
	tabs := container.NewAppTabs(
		container.NewTabItem("Chat", cont),
		container.NewTabItem("Settings", settings),
	)

	tabs.SetTabLocation(container.TabLocationLeading)

	return tabs
}
