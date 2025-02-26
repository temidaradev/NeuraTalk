package internal

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

type Containers interface {
	MakeUI(tabs *container.AppTabs) *fyne.Container
}

type AppTabs interface {
	Sidebar() *container.AppTabs
}
