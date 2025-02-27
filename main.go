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
	w.ShowAndRun()
}
