package internal

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

type Containers interface {
	MakeUI() *fyne.Container
}

type AppTabs interface {
	Sidebar() *container.AppTabs
}
