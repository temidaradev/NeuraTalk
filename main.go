package main

import (
	"fmt"
	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget"
	"github.com/hajimehoshi/guigui/layout"
	"github.com/temidaradev/NeuraTalk/internal"
	"os"
)

type Root struct {
	guigui.RootWidget

	sidebar  internal.Sidebar
	model    internal.Model
	settings internal.Settings
	io       internal.InputOutput

	background basicwidget.Background
}

func (r *Root) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	appender.AppendChildWidgetWithBounds(&r.background, context.Bounds(r))

	r.sidebar.SetModel(&r.model)
	r.io.SetModel(&r.model)

	for i, bounds := range (layout.GridLayout{
		Bounds: context.Bounds(r),
		Widths: []layout.Size{
			layout.FixedSize(8 * basicwidget.UnitSize(context)),
			layout.FlexibleSize(1),
		},
	}).CellBounds() {
		switch i {
		case 0:
			appender.AppendChildWidgetWithBounds(&r.sidebar, bounds)
		case 1:
			switch r.model.Mode() {
			case "io":
				appender.AppendChildWidgetWithBounds(&r.io, bounds)
			case "settings":
				appender.AppendChildWidgetWithBounds(&r.settings, bounds)
			}
		}
	}

	return nil
}

func main() {
	op := &guigui.RunOptions{
		Title: "NeuraTalk ぐいぐい",
	}
	if err := guigui.Run(&Root{}, op); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
