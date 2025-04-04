# NeuraTalk

A modern, user-friendly GUI application for interacting with local Ollama language models. NeuraTalk provides a clean interface for chatting with AI models while offering customization options for a personalized experience.

![NeuraTalk Homepage Light Screenshot](https://github.com/temidaradev/NeuraTalk/blob/ede85a8224671852f3751d0201bd11aaac10fd4c/screenshot_light.png)

![NeuraTalk Homepage Dark Screenshot](https://github.com/temidaradev/NeuraTalk/blob/ede85a8224671852f3751d0201bd11aaac10fd4c/screenshot_dark.png)

## Features

- 🤖 **Local AI Integration**: Seamlessly connect with your local Ollama language models
- 🎨 **Modern Interface**: Clean and intuitive user interface
- 🌓 **Theme Support**: Switch between light and dark themes
- ⚙️ **Customizable Settings**:
  - Adjustable font sizes
  - Configurable response animation speed
  - Auto-scroll toggle
  - Future model-specific settings
- 💬 **Conversation Management**:
  - Clear conversation history
  - Persistent chat history per model
  - Smooth message animations
- ⌨️ **Keyboard Shortcuts**: Quick and efficient interaction
- 🔄 **Cross-Platform Support**: Works on macOS and Linux

## Prerequisites

- Go 1.24.0 or later
- CGO enabled
- Ollama installed and running locally
- For macOS: Homebrew package manager
- For Linux: A supported package manager (apt, dnf, yum, or pacman)

## Installation

### macOS

1. **Clone The Repo**:

   ```bash
   git clone https://github.com/temidaradev/NeuraTalk.git
   cd NeuraTalk
   ```

2. **Make the Script Executable**:

   ```bash
   chmod +x install.sh
   ```

3. **Run the Install Script**:

   ```bash
   ./install.sh
   ```

   The script will:

   - Install required dependencies via Homebrew
   - Build the application
   - Create an application bundle in your Applications folder
   - Set up the application icon

### Linux

1. **Clone The Repo**:

   ```bash
   git clone https://github.com/temidaradev/NeuraTalk.git
   cd NeuraTalk
   ```

2. **Make the Script Executable**:

   ```bash
   chmod +x install.sh
   ```

3. **Run the Install Script**:

   ```bash
   sudo ./install.sh
   ```

   The script will:

   - Install required system dependencies
   - Build the application
   - Install the binary to /usr/local/bin
   - Create desktop entry and icon
   - Set up application menu integration

## Uninstallation

### macOS

1. **Make the Script Executable**:

   ```bash
   chmod +x uninstall.sh
   ```

2. **Run the Uninstall Script**:

   ```bash
   ./uninstall.sh
   ```

   The script will:

   - Remove the application bundle
   - Clean up temporary files
   - Remove conversation history

### Linux

1. **Make the Script Executable**:

   ```bash
   chmod +x uninstall.sh
   ```

2. **Run the Uninstall Script**:

   ```bash
   sudo ./uninstall.sh
   ```

   The script will:

   - Remove the application binary
   - Remove desktop entry and icon
   - Clean up temporary files
   - Remove conversation history

## Usage

1. **Select a Model**:

   - Choose your preferred Ollama model from the dropdown menu at the top
   - Each model maintains its own conversation history

2. **Start Chatting**:

   - Type your message in the input field at the bottom
   - Press Enter to send
   - Watch as the AI responds with a smooth typing animation

3. **Customize Your Experience**:

   - Access settings through the sidebar
   - Adjust theme, font size, and animation speed
   - Toggle auto-scroll behavior

4. **Manage Conversations**:
   - Use the "Clear Conversation" button to start fresh
   - Conversations are automatically saved per model

## Settings

- **Theme**: Switch between light and dark modes
- **Font Size**: Choose between small, medium, and large text
- **Animation Speed**: Adjust the typing animation speed (10-100ms per character)
- **Auto-scroll**: Toggle automatic scrolling to new messages
- **Model Settings**: Configure model-specific parameters (coming soon)

## Development

The project uses:

- [Fyne](https://fyne.io/) for the GUI framework
- [LangChain Go](https://github.com/tmc/langchaingo) for Ollama integration

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the GNU General Public License v3.0 - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Ollama](https://ollama.ai/) for providing the local AI models
- [Fyne](https://fyne.io/) for the excellent GUI framework
- [LangChain Go](https://github.com/tmc/langchaingo) for the Ollama integration
