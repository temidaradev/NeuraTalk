package internal

import (
	"bytes"
	"context"
	"fmt"
	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget"
	"github.com/hajimehoshi/guigui/layout"
	"image"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

type InputOutput struct {
	guigui.RootWidget

	fill         bool
	gap          bool
	Conversation []string

	getOutput     basicwidget.Text
	getInput      basicwidget.TextInput
	sendButton    basicwidget.Button
	selectModel   basicwidget.DropdownList[string]
	scrollArea    basicwidget.ScrollablePanel
	scroll        basicwidget.ScrollOverlay
	selectedModel string
	sidebar       Sidebar

	model    *Model
	settings Settings

	background basicwidget.Background

	animating       bool
	animationTicker *time.Ticker
}

func (io *InputOutput) SetModel(model *Model) {
	io.model = model
}

func (io *InputOutput) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	u := basicwidget.UnitSize(context)
	width := 12 * u

	io.getInput.SetMultiline(true)
	io.getInput.SetAutoWrap(true)
	io.getInput.SetEditable(true)
	context.SetEnabled(&io.getInput, true)

	io.getOutput.SetAutoWrap(true)
	io.getOutput.SetSelectable(true)
	io.getOutput.SetMultiline(true)
	context.SetSize(&io.getOutput, image.Pt(width/2, u))
	io.getInput.SetOnEnterPressed(func(text string) {
		if strings.TrimSpace(text) != "" {
			io.GenerateResponse()
			io.getInput.SetText("")
		}
	})

	for i, bounds := range (layout.GridLayout{
		Bounds: context.Bounds(io).Inset(u / 2),
		Heights: []layout.Size{
			layout.FlexibleSize(1),
			layout.FixedSize(100),
		},
		RowGap: u / 2,
	}).CellBounds() {
		switch i {
		case 0:
			appender.AppendChildWidgetWithBounds(&io.getOutput, bounds)
		case 1:
			appender.AppendChildWidgetWithBounds(&io.getInput, bounds)
		}
	}
	return nil
}

func (io *InputOutput) GetInput() string {
	return io.getInput.Text()
}

func (io *InputOutput) SetOutput(response string) {
	io.Conversation = append(io.Conversation, response)

	io.getOutput.SetText(response)
}

func (io *InputOutput) GenerateResponse() {
	modelName := "llama3.2:latest"

	userPrompt := io.GetInput()
	if strings.TrimSpace(userPrompt) == "" {
		return
	}

	io.getInput.SetEditable(false)

	originalConversation := make([]string, len(io.Conversation))
	copy(originalConversation, io.Conversation)

	thinkingMessage := strings.Join(io.Conversation, "\n\n") + "\n\nYou: " + userPrompt + "\n\nAI: Thinking..."
	io.getOutput.SetText(thinkingMessage)

	go func() {
		ctx := context.Background()

		llm, err := ollama.New(
			ollama.WithModel(modelName),
		)
		if err != nil {
			io.getOutput.SetText(strings.Join(io.Conversation, "\n\n"))
			io.getInput.SetEditable(true)
			return
		}

		fullPrompt := strings.Join(io.Conversation, "\n\n")
		if fullPrompt != "" {
			fullPrompt += "\n\n"
		}
		fullPrompt += userPrompt

		response, err := llms.GenerateFromSinglePrompt(ctx, llm, fullPrompt)
		if err != nil {
			io.Conversation = originalConversation
			io.getOutput.SetText(strings.Join(io.Conversation, "\n\n"))
			io.getInput.SetEditable(true)
			return
		}

		formattedEntry := "You: " + userPrompt + "\n\nAI: " + response
		io.SetOutput(formattedEntry)
		io.getInput.SetEditable(true)
	}()
}

func findOllamaBinary() (string, error) {
	ollamaPath, err := exec.LookPath("ollama")
	if err == nil {
		return ollamaPath, nil
	}

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

	checkCmd := exec.Command(ollamaPath, "list")
	var checkOut bytes.Buffer
	checkCmd.Stdout = &checkOut
	if err := checkCmd.Run(); err != nil {
		return nil, fmt.Errorf("ollama service is not running: %v", err)
	}

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

// GetAvailableModels returns a list of available Ollama models
func GetAvailableModels() ([]string, error) {
	return getOllamaModels()
}
