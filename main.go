package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/creack/pty"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type AppConfig struct {
	AppsDirectory  string
	AllowedApps    map[string]string
	AllowedOrigins map[string]bool
	MaxConcurrent  int
	currentJobs    int
	mu             sync.Mutex
}

type TerminalSession struct {
	conn      *websocket.Conn
	mu        sync.Mutex
	ptmx      *os.File
	cmd       *exec.Cmd
	done      chan bool
	cmdBuffer string
	closed    bool
}

var appConfig = &AppConfig{
	AppsDirectory: "./terminal-apps",
	AllowedApps: map[string]string{
		"tradingcardsearch": "Search for trading cards",
		"testapp":           "App to test if terminal is working when running an app",
	},
	AllowedOrigins: map[string]bool{
		"http://localhost:5173": true,
	},
	MaxConcurrent: 1,
	currentJobs:   0,
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("origin")
		return appConfig.AllowedOrigins[origin]
	},
}

func main() {
	if err := os.MkdirAll(appConfig.AppsDirectory, 0755); err != nil {
		log.Fatalf("Failed to create apps directory: %v", err)
	}

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{
			"http://localhost:5173",
		},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodOptions},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	e.GET("/", handleHome)
	e.GET("/health", handleHealthCheck)
	e.GET("/apps", handleListApps)
	e.GET("/ws", handleWebSocket)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("Terminal Backend Server")
	fmt.Println("Apps Directory: ", appConfig.AppsDirectory)
	fmt.Println("Available Apps: ", getAppsList())
	fmt.Println("Max Concurrent: ", appConfig.MaxConcurrent)
	fmt.Println("Server starting on port:", port)

	e.Logger.Fatal(e.Start(":" + port))
}

func handleHome(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]any{
		"service": "Terminal Backend",
		"version": "1.0.0",
		"status":  "running",
		"endpoints": map[string]string{
			"health":    "/health",
			"apps":      "/apps",
			"websocket": "/ws",
		},
	})
}

func handleHealthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func handleListApps(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]any{
		"apps":      appConfig.AllowedApps,
		"directory": appConfig.AppsDirectory,
	})
}

func handleWebSocket(c echo.Context) error {
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return err
	}
	defer conn.Close()

	log.Printf("New WebSocket connection from: %s", c.Request().RemoteAddr)

	session := &TerminalSession{
		conn:   conn,
		done:   make(chan bool),
		closed: false,
	}
	session.sendWelcome()

	for {
		var msg map[string]any
		err := conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		if command, ok := msg["command"].(string); ok && command != "" {
			session.handleCommand(command)
		} else if input, ok := msg["input"].(string); ok {
			session.handleInput(input)
		} else if resize, ok := msg["resize"].(map[string]any); ok {
			session.handleResize(resize)
		}
	}

	session.cleanup()
	log.Println("WebSocket connection closed")
	return nil
}

func (s *TerminalSession) sendWelcome() {
	welcome := `Welcome to the Terminal Showcase!

Available apps:
`
	for app, desc := range appConfig.AllowedApps {
		welcome += fmt.Sprintf("  %s - %s\n", app, desc)
	}
	welcome += `
Commands:
  <app-name> [args]  - Run an app
  list               - List available apps
  help               - Show this message

`
	s.sendOutput(welcome)
}

func (s *TerminalSession) handleCommand(command string) {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return
	}

	switch parts[0] {
	case "help":
		s.sendWelcome()
		return
	case "list":
		s.listApps()
		return
	case "clear":
		s.sendRawOutput([]byte("\x1b[2J\x1b[H"))
		return
	}

	s.executeApp(parts[0], parts[1:])
}

func (s *TerminalSession) handleInput(input string) {
	s.mu.Lock()
	ptmx := s.ptmx
	s.mu.Unlock()

	log.Printf("Handle Input has been called with: %q, PTMX exists: %v", input, ptmx != nil)
	if ptmx != nil {
		b, err := ptmx.Write([]byte(input))
		if err != nil {
			log.Printf("Error writing to PTY: %v", err)
		} else {
			log.Printf("Wrote %d bytes to PTY.", b)
		}
		return
	}

	log.Printf("No PTY is active at this time. Handling as shell input")

	for _, char := range input {
		switch char {
		case '\r', '\n':
			s.sendRawOutput([]byte("\r\n"))
			if len(s.cmdBuffer) > 0 {
				s.handleCommand(s.cmdBuffer)
				s.cmdBuffer = ""
			}
		case 127, 8:
			if len(s.cmdBuffer) > 0 {
				s.cmdBuffer = s.cmdBuffer[:len(s.cmdBuffer)-1]
				s.sendRawOutput([]byte("\b \b"))
			}
		case 3:
			s.sendRawOutput([]byte("^C\r\n"))
			s.cmdBuffer = ""
		default:
			s.cmdBuffer += string(char)
			s.sendRawOutput([]byte(string(char)))
		}
	}
}

func (s *TerminalSession) listApps() {
	output := "Available apps:\n"
	for app, desc := range appConfig.AllowedApps {
		output += fmt.Sprintf("  %s - %s\n", app, desc)
	}
	s.sendOutput(output)
}

func (s *TerminalSession) executeApp(appName string, args []string) {
	if !appConfig.acquireJob() {
		s.sendOutput("Error: An app is already running. Please wait.\n")
		return
	}
	defer appConfig.releaseJob()

	description, allowed := appConfig.AllowedApps[appName]
	if !allowed {
		s.sendOutput(fmt.Sprintf("Error: App '%s' not found\n", appName))
		s.sendOutput("Type 'list' to see available apps\n")
		return
	}

	appPath := filepath.Join(appConfig.AppsDirectory, appName)

	if _, err := os.Stat(appPath); os.IsNotExist(err) {
		s.sendOutput(fmt.Sprintf("Error: App '%s' executable not found at %s\n", appName, appPath))
		s.sendOutput("Make sure to compile and place your app in the terminal-apps directory\n")
		return
	}

	log.Printf("Running app: %s (%s) with args: %v", appName, description, args)
	s.sendOutput(fmt.Sprintf("Running: %s\n", appName))

	cmd := exec.Command(appPath, args...)

	cmd.Env = append(os.Environ(),
		"TERM=xterm-256color",
		"COLORTERM=truecolor",
		"TERM_PROGRAM=",
	)

	ptmx, err := pty.StartWithSize(cmd, &pty.Winsize{
		Rows: 24,
		Cols: 80,
	})
	if err != nil {
		s.sendOutput(fmt.Sprintf("Error starting app: %v\n", err))
		return
	}

	s.mu.Lock()
	s.ptmx = ptmx
	s.cmd = cmd
	s.mu.Unlock()

	go s.handlePtyOutput(ptmx)

	go func() {
		err = cmd.Wait()

		s.mu.Lock()
		s.ptmx = nil
		s.cmd = nil
		s.mu.Unlock()

		if err != nil {
			log.Printf("App exited with error: %v\n", err)
		}

		s.sendOutput("\r\n[Process Completed. Press Enter to continue]\r\n")
	}()
}

func (s *TerminalSession) handlePtyOutput(ptmx *os.File) {
	buf := make([]byte, 8192)
	for {
		s.mu.Lock()
		closed := s.closed
		s.mu.Unlock()

		if closed {
			return
		}

		n, err := ptmx.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Printf("PTY read error: %v", err)
			}
			return
		}
		if n > 0 {
			s.sendRawOutput(buf[:n])
		}
	}
}

func (s *TerminalSession) handleResize(resize map[string]any) {
	s.mu.Lock()
	ptmx := s.ptmx
	s.mu.Unlock()

	if ptmx == nil {
		return
	}
	rows, rowsOk := resize["rows"].(float64)
	cols, colsOk := resize["cols"].(float64)

	if !rowsOk || !colsOk {
		return
	}

	newRows := uint16(rows)
	newCols := uint16(cols)

	currentSize, err := pty.GetsizeFull(ptmx)
	if err == nil {
		if currentSize.Rows == newRows && currentSize.Cols == newCols {
			return
		}
	}

	err = pty.Setsize(ptmx, &pty.Winsize{
		Rows: uint16(rows),
		Cols: uint16(cols),
	})
	if err != nil {
		log.Printf("Error resizing PTY: %v", err)
	} else {
		log.Printf("Terminal resized to %dx%d", uint16(cols), uint16(rows))
	}
}

func (s *TerminalSession) sendRawOutput(data []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()

	msg := map[string]string{"output": string(data)}
	jsonData, err := json.Marshal(msg)
	if err != nil {
		log.Printf("JSON marshal error: %v", err)
		return
	}

	if err := s.conn.WriteMessage(websocket.TextMessage, jsonData); err != nil {
		log.Printf("Write error: %v", err)
	}
}

func (s *TerminalSession) sendOutput(output string) {
	msg := map[string]string{"output": output}
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("JSON marshal error: %v", err)
		return
	}

	if err := s.conn.WriteMessage(websocket.TextMessage, data); err != nil {
		log.Printf("Write error: %v", err)
	}
}

func (c *AppConfig) acquireJob() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.currentJobs >= c.MaxConcurrent {
		return false
	}
	c.currentJobs++
	log.Printf("App started. Running apps: %d/%d", c.currentJobs, c.MaxConcurrent)
	return true
}

func (s *TerminalSession) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.closed = true

	if s.ptmx != nil {
		s.ptmx.Close()
		s.ptmx = nil
	}

	if s.cmd != nil && s.cmd.Process != nil {
		s.cmd.Process.Kill()
		s.cmd = nil
	}
}

func (c *AppConfig) releaseJob() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.currentJobs > 0 {
		c.currentJobs--
	}
	log.Printf("App finished. Running apps: %d/%d", c.currentJobs, c.MaxConcurrent)
}

func getAppsList() []string {
	apps := []string{}
	for app := range appConfig.AllowedApps {
		apps = append(apps, app)
	}
	return apps
}
