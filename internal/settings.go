package internal

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type Settings struct {
	Window         fyne.Window
	App            fyne.App
	ThemeSelect    *widget.Select
	FontSizeSelect *widget.Select
	AutoScroll     *widget.Check
	AnimationSpeed *widget.Slider
	ModelConfig    *widget.Button
}

func NewSettings(w fyne.Window, a fyne.App) *Settings {
	return &Settings{
		Window: w,
		App:    a,
	}
}

func (s *Settings) GetContainer() *fyne.Container {
	// Theme selection
	themeLabel := widget.NewLabel("Theme:")
	s.ThemeSelect = widget.NewSelect([]string{"Light", "Dark"}, func(selected string) {
		if selected == "Dark" {
			s.App.Settings().SetTheme(theme.DarkTheme())
		} else {
			s.App.Settings().SetTheme(theme.LightTheme())
		}
	})
	s.ThemeSelect.SetSelected("Light")

	// Font size selection
	fontLabel := widget.NewLabel("Font Size:")
	s.FontSizeSelect = widget.NewSelect([]string{"Small", "Medium", "Large"}, func(selected string) {
		// Font size changes will be handled by the main app
	})

	// Auto-scroll toggle
	s.AutoScroll = widget.NewCheck("Auto-scroll to new messages", func(checked bool) {
		// Auto-scroll behavior will be handled by the main app
	})
	s.AutoScroll.SetChecked(true)

	// Animation speed slider
	animationLabel := widget.NewLabel("Response Animation Speed:")
	s.AnimationSpeed = widget.NewSlider(10, 100)
	s.AnimationSpeed.SetValue(20) // Default to 20ms per character

	// Model configuration button
	s.ModelConfig = widget.NewButton("Configure Model Settings", func() {
		// Model configuration dialog will be shown
		dialog.ShowInformation("Model Configuration",
			"Model configuration options will be available in a future update.",
			s.Window)
	})

	// Create a form layout for settings
	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Theme", Widget: container.NewHBox(themeLabel, s.ThemeSelect)},
			{Text: "Font Size", Widget: container.NewHBox(fontLabel, s.FontSizeSelect)},
			{Text: "", Widget: s.AutoScroll},
			{Text: "Animation Speed", Widget: container.NewVBox(
				animationLabel,
				s.AnimationSpeed,
				widget.NewLabel("(slower) ← → (faster)"),
			)},
		},
		OnSubmit: func() {
			// Save settings if needed
		},
	}

	// Create a container with the form and model config button
	content := container.NewVBox(
		widget.NewLabel("Settings"),
		widget.NewSeparator(),
		form,
		widget.NewSeparator(),
		s.ModelConfig,
	)

	return container.NewBorder(
		nil,     // top
		nil,     // bottom
		nil,     // left
		nil,     // right
		content, // center
	)
}

// GetTheme returns the current theme selection
func (s *Settings) GetTheme() string {
	return s.ThemeSelect.Selected
}

// GetFontSize returns the current font size selection
func (s *Settings) GetFontSize() string {
	return s.FontSizeSelect.Selected
}

// IsAutoScrollEnabled returns whether auto-scroll is enabled
func (s *Settings) IsAutoScrollEnabled() bool {
	return s.AutoScroll.Checked
}

// GetAnimationSpeed returns the current animation speed value
func (s *Settings) GetAnimationSpeed() float64 {
	return s.AnimationSpeed.Value
}
