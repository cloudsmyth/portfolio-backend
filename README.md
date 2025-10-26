# Terminal Backend Server

A WebSocket-based terminal emulation backend server written in Go that allows running approved terminal applications through a web interface.

## Overview

This server provides a WebSocket API for executing and interacting with terminal applications in a controlled environment. It uses pseudo-terminals (PTY) to provide full terminal emulation, supporting ANSI escape codes, colors, and interactive input/output.

## Features

- **WebSocket-based terminal emulation** - Full bidirectional communication with PTY support
- **Application sandboxing** - Only whitelisted applications can be executed
- **Concurrent job control** - Configurable limit on simultaneously running applications
- **Terminal resize support** - Dynamic terminal window resizing
- **CORS protection** - Configurable allowed origins
- **Health monitoring** - Built-in health check endpoint

## Architecture

### Main Components

- **AppConfig**: Global configuration managing allowed apps, origins, and concurrency limits
- **TerminalSession**: Individual WebSocket connection handler managing PTY and command execution
- **WebSocket Handler**: Manages bidirectional communication between client and terminal

### Directory Structure

```
.
├── main.go                    # Main server code
├── terminal-apps-exe/         # Compiled executables directory
│   ├── tradingcardsearch
│   └── testapp
└── terminal-apps/             # Source code submodules
    └── [submodule directories]
```

## API Endpoints

### HTTP Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/` | GET | Service information and available endpoints |
| `/health` | GET | Health check endpoint |
| `/apps` | GET | List available applications |
| `/ws` | GET | WebSocket upgrade endpoint |

### WebSocket Protocol

The WebSocket connection accepts JSON messages with the following formats:

#### Client to Server

**Execute Command:**
```json
{
  "command": "appname [args]"
}
```

**Send Input to Running App:**
```json
{
  "input": "user input text"
}
```

**Resize Terminal:**
```json
{
  "resize": {
    "rows": 24,
    "cols": 80
  }
}
```

#### Server to Client

**Terminal Output:**
```json
{
  "output": "terminal output data"
}
```

## Built-in Commands

- `help` - Display welcome message and available apps
- `list` - List all available applications
- `clear` - Clear the terminal screen
- `<app-name> [args]` - Execute a whitelisted application

## Configuration

### Environment Variables

- `PORT` - Server port (default: `8080`)

## Security Considerations

- Only whitelisted applications can be executed
- CORS protection limits allowed origins
- Applications run with the same permissions as the server process
- Consider running the server in a containerized environment
- Limit concurrent executions to prevent resource exhaustion

## Terminal Features

The PTY implementation supports:

- **ANSI escape codes** for colors and formatting
- **Interactive input/output** with stdin/stdout/stderr
- **Terminal resizing** with proper signal handling
- **256-color and truecolor** support
- **Control characters** (Ctrl+C, backspace, etc.)

## Logging

The server logs:
- WebSocket connection events
- Application execution and termination
- Error conditions
- Concurrent job status

## Error Handling

- Invalid app names return error messages
- Missing executables provide helpful feedback
- Concurrent execution limits are enforced
- WebSocket errors are logged and handled gracefully

## License

[Add your license information here]

## Contributing

[Add contribution guidelines here]
