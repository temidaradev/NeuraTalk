package internal

import (
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type Sidebar struct{}

func (s *Sidebar) Sidebar() *container.AppTabs {
	aiResponseArea := CreateAIResponseArea()
	tabs := container.NewAppTabs(
		container.NewTabItem("AI Response", aiResponseArea),
		container.NewTabItem("Tab 2", widget.NewLabel("World!")),
	)

	tabs.SetTabLocation(container.TabLocationLeading)

	return tabs
}
