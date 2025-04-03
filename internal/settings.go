package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// SettingsData represents the structure of saved settings
type SettingsData struct {
	Theme          string  `json:"theme"`
	FontSize       string  `json:"fontSize"`
	AutoScroll     bool    `json:"autoScroll"`
	AnimationSpeed float64 `json:"animationSpeed"`
	Model          string  `json:"model"`
	Temperature    float64 `json:"temperature"`
	MaxTokens      float64 `json:"maxTokens"`
	TopP           float64 `json:"topP"`
	TopK           float64 `json:"topK"`
	ContextLength  float64 `json:"contextLength"`
}

type Settings struct {
	Window         fyne.Window
	App            fyne.App
	ThemeSelect    *widget.Select
	FontSizeSelect *widget.Select
	AutoScroll     *widget.Check
	AnimationSpeed *widget.Slider
	ModelConfig    *widget.Button
	// LLM Settings
	TemperatureSlider   *widget.Slider
	MaxTokensSlider     *widget.Slider
	TopPSlider          *widget.Slider
	TopKSlider          *widget.Slider
	ContextLengthSlider *widget.Slider
	ModelSelect         *widget.Select
}

func NewSettings(w fyne.Window, a fyne.App) *Settings {
	s := &Settings{
		Window: w,
		App:    a,
	}

	// Initialize UI elements first
	s.initializeUIElements()

	// Then load saved settings
	s.loadSettings()

	return s
}

func (s *Settings) initializeUIElements() {
	// Theme selection
	s.ThemeSelect = widget.NewSelect([]string{"Light", "Dark"}, func(selected string) {
		if selected == "Dark" {
			s.App.Settings().SetTheme(theme.DarkTheme())
		} else {
			s.App.Settings().SetTheme(theme.LightTheme())
		}
		s.saveSettings()
	})

	// Font size selection
	s.FontSizeSelect = widget.NewSelect([]string{"Small", "Medium", "Large"}, func(selected string) {
		s.saveSettings()
	})

	// Auto-scroll toggle
	s.AutoScroll = widget.NewCheck("Auto-scroll to new messages", func(checked bool) {
		s.saveSettings()
	})

	// Animation speed slider
	s.AnimationSpeed = widget.NewSlider(10, 100)
	s.AnimationSpeed.OnChanged = func(value float64) {
		s.saveSettings()
	}

	// Model selection
	s.ModelSelect = widget.NewSelect([]string{}, func(selected string) {
		s.saveSettings()
	})

	// Temperature slider
	s.TemperatureSlider = widget.NewSlider(0, 2)
	s.TemperatureSlider.OnChanged = func(value float64) {
		s.saveSettings()
	}

	// Max tokens slider
	s.MaxTokensSlider = widget.NewSlider(100, 4096)
	s.MaxTokensSlider.OnChanged = func(value float64) {
		s.saveSettings()
	}

	// Top P slider
	s.TopPSlider = widget.NewSlider(0, 1)
	s.TopPSlider.OnChanged = func(value float64) {
		s.saveSettings()
	}

	// Top K slider
	s.TopKSlider = widget.NewSlider(1, 100)
	s.TopKSlider.OnChanged = func(value float64) {
		s.saveSettings()
	}

	// Context length slider
	s.ContextLengthSlider = widget.NewSlider(512, 8192)
	s.ContextLengthSlider.OnChanged = func(value float64) {
		s.saveSettings()
	}
}

func (s *Settings) saveSettings() {
	// Create config directory if it doesn't exist
	configDir := filepath.Join(".", "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		dialog.ShowError(fmt.Errorf("Failed to create config directory: %v", err), s.Window)
		return
	}

	// Read existing settings if any
	configPath := filepath.Join(configDir, "settings.json")
	var existingSettings SettingsData
	if data, err := os.ReadFile(configPath); err == nil {
		if err := json.Unmarshal(data, &existingSettings); err != nil {
			// If we can't read existing settings, start fresh
			existingSettings = SettingsData{}
		}
	}

	// Update only the changed settings
	settings := SettingsData{
		Theme:          s.ThemeSelect.Selected,
		FontSize:       s.FontSizeSelect.Selected,
		AutoScroll:     s.AutoScroll.Checked,
		AnimationSpeed: s.AnimationSpeed.Value,
		Model:          s.ModelSelect.Selected,
		Temperature:    s.TemperatureSlider.Value,
		MaxTokens:      s.MaxTokensSlider.Value,
		TopP:           s.TopPSlider.Value,
		TopK:           s.TopKSlider.Value,
		ContextLength:  s.ContextLengthSlider.Value,
	}

	// Save settings to file with pretty formatting
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		dialog.ShowError(fmt.Errorf("Failed to marshal settings: %v", err), s.Window)
		return
	}

	// Write to a temporary file first
	tempPath := configPath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		dialog.ShowError(fmt.Errorf("Failed to write temporary settings file: %v", err), s.Window)
		return
	}

	// Rename temporary file to actual file (atomic operation)
	if err := os.Rename(tempPath, configPath); err != nil {
		dialog.ShowError(fmt.Errorf("Failed to save settings: %v", err), s.Window)
		// Clean up temp file if it exists
		os.Remove(tempPath)
		return
	}
}

func (s *Settings) loadSettings() {
	// Default settings
	defaultSettings := SettingsData{
		Theme:          "Light",
		FontSize:       "Medium",
		AutoScroll:     true,
		AnimationSpeed: 20,
		Model:          "",
		Temperature:    0.7,
		MaxTokens:      2048,
		TopP:           0.9,
		TopK:           40,
		ContextLength:  4096,
	}

	// Try to load saved settings
	configPath := filepath.Join(".", "config", "settings.json")
	data, err := os.ReadFile(configPath)
	if err == nil {
		var settings SettingsData
		if err := json.Unmarshal(data, &settings); err == nil {
			defaultSettings = settings
		}
	}

	// Apply loaded settings to UI elements
	if s.ThemeSelect != nil {
		s.ThemeSelect.SetSelected(defaultSettings.Theme)
		// Apply theme immediately
		if defaultSettings.Theme == "Dark" {
			s.App.Settings().SetTheme(theme.DarkTheme())
		} else {
			s.App.Settings().SetTheme(theme.LightTheme())
		}
	}

	if s.FontSizeSelect != nil {
		s.FontSizeSelect.SetSelected(defaultSettings.FontSize)
	}

	if s.AutoScroll != nil {
		s.AutoScroll.SetChecked(defaultSettings.AutoScroll)
	}

	if s.AnimationSpeed != nil {
		s.AnimationSpeed.SetValue(defaultSettings.AnimationSpeed)
	}

	if s.ModelSelect != nil {
		s.ModelSelect.SetSelected(defaultSettings.Model)
	}

	if s.TemperatureSlider != nil {
		s.TemperatureSlider.SetValue(defaultSettings.Temperature)
	}

	if s.MaxTokensSlider != nil {
		s.MaxTokensSlider.SetValue(defaultSettings.MaxTokens)
	}

	if s.TopPSlider != nil {
		s.TopPSlider.SetValue(defaultSettings.TopP)
	}

	if s.TopKSlider != nil {
		s.TopKSlider.SetValue(defaultSettings.TopK)
	}

	if s.ContextLengthSlider != nil {
		s.ContextLengthSlider.SetValue(defaultSettings.ContextLength)
	}
}

func (s *Settings) GetContainer() *fyne.Container {
	// Create labels for the settings
	themeLabel := widget.NewLabel("Theme:")
	fontLabel := widget.NewLabel("Font Size:")
	llmSettingsLabel := widget.NewLabelWithStyle("LLM Settings", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	modelLabel := widget.NewLabel("Model:")

	// Create a container with the form and model config button
	content := container.NewVBox(
		widget.NewLabel("Settings"),
		widget.NewSeparator(),
		container.NewVBox(
			widget.NewLabel("Theme"),
			container.NewHBox(themeLabel, s.ThemeSelect),
			widget.NewLabel("Font Size"),
			container.NewHBox(fontLabel, s.FontSizeSelect),
			s.AutoScroll,
			widget.NewLabel("Animation Speed"),
			s.AnimationSpeed,
			widget.NewLabel("(slower) ← → (faster)"),
		),
		widget.NewSeparator(),
		llmSettingsLabel,
		widget.NewSeparator(),
		container.NewVBox(
			widget.NewLabel("Model"),
			container.NewHBox(modelLabel, s.ModelSelect),
			widget.NewLabel("Temperature"),
			s.TemperatureSlider,
			widget.NewLabel("(deterministic) ← → (creative)"),
			widget.NewLabel("Max Tokens"),
			s.MaxTokensSlider,
			widget.NewLabel("(shorter) ← → (longer)"),
			widget.NewLabel("Top P"),
			s.TopPSlider,
			widget.NewLabel("(focused) ← → (diverse)"),
			widget.NewLabel("Top K"),
			s.TopKSlider,
			widget.NewLabel("(focused) ← → (diverse)"),
			widget.NewLabel("Context Length"),
			s.ContextLengthSlider,
			widget.NewLabel("(shorter) ← → (longer)"),
		),
	)

	// Wrap the content in a scroll container
	scrollContent := container.NewVScroll(content)
	scrollContent.SetMinSize(fyne.NewSize(300, 400))

	return container.NewBorder(
		nil,           // top
		nil,           // bottom
		nil,           // left
		nil,           // right
		scrollContent, // center
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

// LLM Settings getters
func (s *Settings) GetModel() string {
	return s.ModelSelect.Selected
}

func (s *Settings) GetTemperature() float64 {
	return s.TemperatureSlider.Value
}

func (s *Settings) GetMaxTokens() float64 {
	return s.MaxTokensSlider.Value
}

func (s *Settings) GetTopP() float64 {
	return s.TopPSlider.Value
}

func (s *Settings) GetTopK() float64 {
	return s.TopKSlider.Value
}

func (s *Settings) GetContextLength() float64 {
	return s.ContextLengthSlider.Value
}
