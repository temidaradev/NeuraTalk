package main

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/temidaradev/NeuraTalk/internal"
)

func main() {
	a := app.New()
	w := a.NewWindow("NeuraTalk")
	w.Resize(fyne.NewSize(800, 600))

	cmd := exec.Command("ollama", "list")
	cmd.Dir = "/home"
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	output := out.String()
	fmt.Println(output)

	lines := strings.Split(output, "\n")
	var names []string
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

	fmt.Println("Parsed Names:", names)

	io := internal.NewInputOutput(names, w)

	var apptabs internal.AppTabs = &internal.Sidebar{}
	tabs := apptabs.Sidebar(io.GetContainer())

	var drawable internal.Containers = &internal.UI{}
	ui := drawable.MakeUI(tabs)

	w.SetContent(ui)

	// selection := internal.NewSelection()

	// selectModelButton := selection.SelectButtonCallback(w)

	// // Content for the pop-up window
	// popupContent := container.NewVBox(
	// 	selection.GetContainer(),
	// 	selectModelButton,
	// )

	// // Show startup pop-up screen
	// startupContent := container.NewVBox(
	// 	widget.NewLabel("Welcome to NeuraTalk!"),
	// 	widget.NewButton("Get Started", func() {
	// 		dialog.ShowCustom("Select Model", "Close", popupContent, w)
	// 	}),
	// )

	// startupDialog := dialog.NewCustom("Welcome", "Close", startupContent, w)
	// startupDialog.Show()

	// // Main window content
	// mainContent := container.NewVBox(
	// 	widget.NewLabel("Main Application Content"),
	// 	widget.NewButton("Send", func() {
	// 		modelName := selection.SelectedModel
	// 		if modelName == "" {
	// 			dialog.ShowInformation("Error", "Please select a model first.", w)
	// 			return
	// 		}

	// 		ctx := context.Background()
	// 		llm, err := ollama.New(ollama.WithModel(modelName))
	// 		if err != nil {
	// 			dialog.ShowError(err, w)
	// 			return
	// 		}

	// 		prompt := "What would be a good company name for a company that makes colorful socks?"
	// 		response, err := llms.GenerateFromSinglePrompt(ctx, llm, prompt)
	// 		if err != nil {
	// 			dialog.ShowError(err, w)
	// 			return
	// 		}

	// 		dialog.ShowInformation("AI Response", response, w)
	// 	}),
	// )

	// w.SetContent(mainContent)
	w.ShowAndRun()
}
