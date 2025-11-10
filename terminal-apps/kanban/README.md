# Kanban CLI

A terminal-based kanban board application built with Go and the Bubble Tea TUI framework. Manage your tasks and workflow directly from your terminal with an intuitive, keyboard-driven interface.

## Features

- **Terminal-Native Interface**: Built with Bubble Tea for a smooth, responsive TUI experience
- **Kanban Workflow**: Classic board layout with Todo, In Progress, and Done columns
- **Keyboard-Driven**: Navigate and manage tasks efficiently without touching your mouse
- **Lightweight**: Fast startup and minimal resource usage
- **Cross-Platform**: Works on Linux, macOS, and Windows

## Installation

### From Source

Requires Go 1.21 or later:

```bash
git clone https://github.com/yourusername/kanban-go.git
cd kanban-go
go build -o kanban
```

### Using Go Install

```bash
go install github.com/yourusername/kanban-go@latest
```

## Usage

Start the application:

```bash
./kanban
```

### Keyboard Shortcuts

#### Navigation
- `←/→` or `h/l` - Move between columns
- `↑/↓` or `j/k` - Move between tasks in a column
- `tab` - Cycle through columns

#### Task Management
- `n` - Create a new task
- `e` - Edit selected task
- `d` - Delete selected task
- `enter` - Move task to next column
- `backspace` - Move task to previous column

#### Application
- `?` - Toggle help menu
- `q` or `ctrl+c` - Quit application

## Project Structure

```
kanban-go/
├── main.go           # Application entry point
├── model.go          # Bubble Tea model and state
├── view.go           # UI rendering logic
├── update.go         # Event handling and state updates
└── go.mod            # Dependencies
```

## Technology Stack

- **[Bubble Tea](https://github.com/charmbracelet/bubbletea)**: TUI framework following The Elm Architecture
- **[Lipgloss](https://github.com/charmbracelet/lipgloss)**: Terminal styling and layout
- **[Bubbles](https://github.com/charmbracelet/bubbles)**: Reusable TUI components

## Roadmap

- [ ] Task priority levels with visual indicators
- [ ] Due dates and reminders
- [ ] Task filtering and search
- [ ] Multiple board support
- [ ] Task tags and categories
- [ ] Export to Markdown/CSV
- [ ] Sync with external services (GitHub Issues, Jira)
- [ ] Customizable columns
- [ ] Task time tracking
- [ ] Vim-style command mode

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- Built with the excellent [Charm](https://charm.sh/) suite of TUI tools
- Inspired by the simplicity of physical kanban boards
