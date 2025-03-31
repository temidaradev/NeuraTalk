package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/temidaradev/NeuraTalk/internal"
)

func findOllamaBinary() (string, error) {
	// Check if ollama is in PATH
	ollamaPath, err := exec.LookPath("ollama")
	if err == nil {
		return ollamaPath, nil
	}

	// Common installation paths based on OS
	var possiblePaths []string
	switch runtime.GOOS {
	case "darwin":
		possiblePaths = []string{
			"/usr/local/bin/ollama",
			"/opt/homebrew/bin/ollama",
			filepath.Join(os.Getenv("HOME"), "go/bin/ollama"),
		}
	case "windows":
		possiblePaths = []string{
			"C:\\Program Files\\Ollama\\ollama.exe",
			"C:\\Program Files (x86)\\Ollama\\ollama.exe",
			filepath.Join(os.Getenv("LOCALAPPDATA"), "ollama\\ollama.exe"),
		}
	case "linux":
		possiblePaths = []string{
			"/usr/bin/ollama",
			"/usr/local/bin/ollama",
			filepath.Join(os.Getenv("HOME"), "go/bin/ollama"),
		}
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("ollama not found in PATH or common installation locations")
}

func getOllamaModels() ([]string, error) {
	ollamaPath, err := findOllamaBinary()
	if err != nil {
		return nil, fmt.Errorf("failed to find ollama: %v", err)
	}

	// Check if ollama service is running
	checkCmd := exec.Command(ollamaPath, "list")
	var checkOut bytes.Buffer
	checkCmd.Stdout = &checkOut
	if err := checkCmd.Run(); err != nil {
		return nil, fmt.Errorf("ollama service is not running: %v", err)
	}

	// Get list of models
	cmd := exec.Command(ollamaPath, "list")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to list models: %v", err)
	}

	output := out.String()
	var names []string
	lines := strings.Split(output, "\n")
	startParsing := false
	for _, line := range lines {
		if strings.HasPrefix(line, "NAME") {
			startParsing = true
			continue
		}
		if startParsing {
			columns := strings.Fields(line)
			if len(columns) > 0 {
				names = append(names, columns[0])
			}
		}
	}

	if len(names) == 0 {
		return nil, fmt.Errorf("no models found. Please install a model using 'ollama pull <model-name>'")
	}

	return names, nil
}

type ChatManager struct {
	Instances []*internal.InputOutput
	Current   int
	Sidebar   *internal.Sidebar
	Window    fyne.Window
	Settings  *internal.Settings
}

func NewChatManager(names []string, w fyne.Window, a fyne.App) *ChatManager {
	manager := &ChatManager{
		Instances: make([]*internal.InputOutput, 0),
		Current:   0,
		Window:    w,
	}

	// Create settings first
	manager.Settings = internal.NewSettings(w, a)

	// Create sidebar
	manager.Sidebar = &internal.Sidebar{}

	// Set up new chat functionality
	newChatFunc := func() {
		// Create a new chat instance
		newIO := internal.NewInputOutput(names, w)
		manager.Instances = append(manager.Instances, newIO)
		manager.Current = len(manager.Instances) - 1

		// Add new chat tab and switch to it
		chatTab := container.NewTabItemWithIcon(
			fmt.Sprintf("Chat %d", len(manager.Instances)),
			theme.DocumentIcon(),
			newIO.GetContainer(),
		)
		manager.Sidebar.TabContainer.Append(chatTab)
		manager.Sidebar.TabContainer.Select(chatTab)
	}

	// Create the New Chat button with the functionality
	manager.Sidebar.NewChatButton = widget.NewButtonWithIcon("New Chat", theme.ContentAddIcon(), newChatFunc)

	return manager
}

func main() {
	a := app.New()
	w := a.NewWindow("NeuraTalk")
	w.Resize(fyne.NewSize(800, 600))

	// Get list of available models
	names, err := getOllamaModels()
	if err != nil {
		dialog.ShowError(err, w)
		// Show instructions for installing Ollama
		instructions := "To use NeuraTalk, you need to:\n\n" +
			"1. Install Ollama from https://ollama.ai\n" +
			"2. Start the Ollama service with 'ollama serve'\n" +
			"3. Install a model with 'ollama pull <model-name>'\n\n" +
			"Example: ollama pull llama2"
		dialog.ShowInformation("Setup Required", instructions, w)
		return
	}

	fmt.Println("Available Models:", names)

	// Create chat manager with settings
	manager := NewChatManager(names, w, a)

	// Create initial UI
	split := manager.Sidebar.Sidebar(nil, manager.Settings.GetContainer())
	w.SetContent(split)
	w.ShowAndRun()
}
