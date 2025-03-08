package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/dialog"
	"github.com/temidaradev/NeuraTalk/internal"
)

func main() {
	a := app.New()
	w := a.NewWindow("NeuraTalk")
	w.Resize(fyne.NewSize(800, 600))

	var names []string
	cmd := exec.Command("ollama", "list")
	cmd.Dir = "/home"
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		dialog.ShowInformation("Important!", "Ollama not installed or failed to run", w)
	} else {
		output := out.String()
		fmt.Println(output)

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
	}

	fmt.Println("Parsed Names:", names)

	io := internal.NewInputOutput(names, w)
	s := internal.NewSettings()

	var apptabs internal.AppTabs = &internal.Sidebar{}
	tabs := apptabs.Sidebar(io.GetContainer(), s.GetContainer())

	var drawable internal.Containers = &internal.UI{}
	ui := drawable.MakeUI(tabs)

	w.SetContent(ui)
	w.ShowAndRun()
}
