package internal

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type Sidebar struct {
	NewChatButton  *widget.Button
	LastChatButton *widget.Button
	OptionsButton  *widget.Button
	HomeButton     *widget.Button
	TabContainer   *container.DocTabs
	MainContent    *fyne.Container
}

func (s *Sidebar) Sidebar(cont *fyne.Container, settings *fyne.Container) *container.Split {
	// Create main content area with closeable tabs
	s.TabContainer = container.NewDocTabs()
	s.MainContent = cont

	// Create welcome content
	welcomeTitle := widget.NewLabelWithStyle("Welcome to NeuraTalk", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	welcomeTitle.TextStyle.Bold = true

	welcomeText := widget.NewTextGridFromString(
		"Your AI companion for seamless conversations.\n\n" +
			"Start a new chat to begin your journey!")

	// Create Last Chat button for the home menu
	welcomeLastChatBtn := widget.NewButtonWithIcon("Last Chat", theme.DocumentIcon(), func() {
		if s.LastChatButton != nil && s.LastChatButton.OnTapped != nil {
			s.LastChatButton.OnTapped()
		}
	})

	welcomeNewChatBtn := widget.NewButtonWithIcon("New Chat", theme.ContentAddIcon(), func() {
		if s.NewChatButton != nil && s.NewChatButton.OnTapped != nil {
			s.NewChatButton.OnTapped()
		}
	})

	welcomeContent := container.NewVBox(
		layout.NewSpacer(),
		welcomeTitle,
		welcomeText,
		layout.NewSpacer(),
		container.NewHBox(
			layout.NewSpacer(),
			welcomeLastChatBtn,
			layout.NewSpacer(),
		),
		container.NewHBox(
			layout.NewSpacer(),
			welcomeNewChatBtn,
			layout.NewSpacer(),
		),
		layout.NewSpacer(),
	)

	// Create home button
	s.HomeButton = widget.NewButtonWithIcon("Home", theme.HomeIcon(), func() {
		// Check if home tab already exists
		for _, tab := range s.TabContainer.Items {
			if tab.Text == "Home" {
				s.TabContainer.Select(tab)
				return
			}
		}
		// Create new home tab
		homeTab := container.NewTabItemWithIcon("Home", theme.HomeIcon(), welcomeContent)
		s.TabContainer.Append(homeTab)
		s.TabContainer.Select(homeTab)
	})

	// Create options button
	s.OptionsButton = widget.NewButtonWithIcon("Options", theme.SettingsIcon(), func() {
		// Check if options tab already exists
		for _, tab := range s.TabContainer.Items {
			if tab.Text == "Options" {
				s.TabContainer.Select(tab)
				return
			}
		}
		// Create new options tab
		optionsTab := container.NewTabItemWithIcon("Options", theme.SettingsIcon(), settings)
		s.TabContainer.Append(optionsTab)
		s.TabContainer.Select(optionsTab)
	})

	// Create sidebar content
	topContent := container.NewVBox(
		s.HomeButton,
		widget.NewSeparator(),
		s.NewChatButton,
		widget.NewSeparator(),
		s.OptionsButton,
	)

	// Add padding around the buttons
	paddedContent := container.NewPadded(topContent)

	// Create a split container with resizable sidebar
	split := container.NewHSplit(
		paddedContent,
		s.TabContainer,
	)
	split.SetOffset(0.2) // Set initial sidebar width to 20% of window width

	// Show home tab by default
	s.HomeButton.OnTapped()

	return split
}
