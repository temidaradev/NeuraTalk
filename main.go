package main

import (
	"fmt"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/temidaradev/NeuraTalk/internal"
)

type ChatManager struct {
	Instances []*internal.InputOutput
	Current   int
	Window    fyne.Window
	Sidebar   *internal.Sidebar
	Settings  *internal.Settings
	LastChat  *internal.InputOutput
}

func NewChatManager(w fyne.Window, settings *internal.Settings) *ChatManager {
	// Get available models
	models, err := internal.GetAvailableModels()
	if err != nil {
		dialog.ShowError(fmt.Errorf("Failed to get available models: %v", err), w)
		return nil
	}

	// Create first chat instance
	io := internal.NewInputOutput(models, w, settings)

	manager := &ChatManager{
		Instances: []*internal.InputOutput{io},
		Current:   0,
		Window:    w,
		Settings:  settings,
		LastChat:  io,
	}

	// Create sidebar
	manager.Sidebar = &internal.Sidebar{}

	// Set up new chat functionality
	newChatFunc := func() {
		// Create a new chat instance
		newIO := internal.NewInputOutput(models, w, settings)
		manager.Instances = append(manager.Instances, newIO)
		manager.Current = len(manager.Instances) - 1
		manager.LastChat = newIO

		// Add new chat tab and switch to it
		chatTab := container.NewTabItemWithIcon(
			fmt.Sprintf("Chat %d", len(manager.Instances)),
			theme.DocumentIcon(),
			newIO.GetContainer(),
		)
		manager.Sidebar.TabContainer.Append(chatTab)
		manager.Sidebar.TabContainer.Select(chatTab)
	}

	// Set up last chat functionality
	lastChatFunc := func() {
		if manager.LastChat != nil {
			// Check if the last chat tab already exists
			for _, tab := range manager.Sidebar.TabContainer.Items {
				if tab.Content == manager.LastChat.GetContainer() {
					manager.Sidebar.TabContainer.Select(tab)
					return
				}
			}

			// Create new tab for last chat
			chatTab := container.NewTabItemWithIcon(
				"Last Chat",
				theme.DocumentIcon(),
				manager.LastChat.GetContainer(),
			)
			manager.Sidebar.TabContainer.Append(chatTab)
			manager.Sidebar.TabContainer.Select(chatTab)
		}
	}

	// Create the New Chat button with the functionality
	manager.Sidebar.NewChatButton = widget.NewButtonWithIcon("New Chat", theme.ContentAddIcon(), newChatFunc)
	manager.Sidebar.LastChatButton = widget.NewButtonWithIcon("Last Chat", theme.DocumentIcon(), lastChatFunc)

	return manager
}

// SetLastChat updates the last chat instance
func (m *ChatManager) SetLastChat(io *internal.InputOutput) {
	m.LastChat = io
	m.Instances = append(m.Instances, io)
	m.Current = len(m.Instances) - 1

	// Add new chat tab and switch to it
	chatTab := container.NewTabItemWithIcon(
		"Last Chat",
		theme.DocumentIcon(),
		io.GetContainer(),
	)
	m.Sidebar.TabContainer.Append(chatTab)
	m.Sidebar.TabContainer.Select(chatTab)
}

func main() {
	a := app.New()
	w := a.NewWindow("NeuraTalk")
	w.Resize(fyne.NewSize(800, 600))

	// Set initial theme before creating any widgets
	if settings := a.Settings(); settings != nil {
		settings.SetTheme(theme.LightTheme())
	}

	// Create settings
	settings := internal.NewSettings(w, a)

	// Create chat manager with settings
	manager := NewChatManager(w, settings)

	// Get list of available models
	names, err := internal.GetAvailableModels()
	if err != nil {
		dialog.ShowError(fmt.Errorf("Failed to get available models: %v", err), w)
		os.Exit(1)
	}

	fmt.Println("Available Models:", names)

	// Create initial UI
	split := manager.Sidebar.Sidebar(nil, settings.GetContainer())
	w.SetContent(split)

	// Start with the last chat if it exists
	if manager.LastChat != nil {
		// Create a new tab for the last chat
		chatTab := container.NewTabItemWithIcon(
			"Last Chat",
			theme.DocumentIcon(),
			manager.LastChat.GetContainer(),
		)
		manager.Sidebar.TabContainer.Append(chatTab)
		manager.Sidebar.TabContainer.Select(chatTab)
	}

	// Center the window on screen
	w.CenterOnScreen()

	// Show and run
	w.ShowAndRun()
}
