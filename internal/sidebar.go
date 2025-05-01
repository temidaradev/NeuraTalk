package internal

import (
	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget"
)

type Sidebar struct {
	guigui.DefaultWidget

	sidebar        basicwidget.Sidebar
	sidebarContent SidebarContent
}

func (s *Sidebar) SetModel(model *Model) {
	s.sidebarContent.SetModel(model)
}

func (s *Sidebar) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	context.SetSize(&s.sidebarContent, context.Size(s))
	s.sidebar.SetContent(&s.sidebarContent)

	appender.AppendChildWidgetWithBounds(&s.sidebar, context.Bounds(s))

	return nil
}

type SidebarContent struct {
	guigui.DefaultWidget
	list basicwidget.TextList[string]

	model *Model
}

func (s *SidebarContent) SetModel(model *Model) {
	s.model = model
}

func (s *SidebarContent) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	s.list.SetStyle(basicwidget.ListStyleSidebar)

	items := []basicwidget.TextListItem[string]{
		{
			Text: "Home",
			Tag:  "io",
		},
		{
			Text: "Settings",
			Tag:  "settings",
		},
	}

	s.list.SetItems(items)
	s.list.SelectItemByTag(s.model.Mode())
	s.list.SetItemHeight(basicwidget.UnitSize(context))
	s.list.SetOnItemSelected(func(index int) {
		item, ok := s.list.ItemByIndex(index)
		if !ok {
			s.model.SetMode("")
			return
		}
		s.model.SetMode(item.Tag)
	})

	appender.AppendChildWidgetWithBounds(&s.list, context.Bounds(s))

	return nil
}
