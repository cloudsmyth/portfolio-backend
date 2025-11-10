package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/cloudsmyth/portfolio-backend/handlers"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	godotenv.Load()

	terminalConfig := &handlers.TerminalConfig{
		AppsDirectory: "./terminal-apps-exe",
		AllowedApps: map[string]string{
			"tradingcardsearch": "Search for trading cards",
			"testapp":           "App to test if terminal is working when running an app",
			"kanban":            "Classic kanban style app",
		},
		AllowedOrigins: map[string]bool{
			"http://localhost:5173":       true,
			"https://spenceralan.dev":     true,
			"https://www.spenceralan.dev": true,
		},
		MaxConcurrent: 1,
	}

	if err := os.MkdirAll(terminalConfig.AppsDirectory, 0755); err != nil {
		log.Fatalf("Failed to create apps directory: %v", err)
	}

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{
			"http://localhost:5173",
			"https://spenceralan.dev",
			"https://www.spenceralan.dev",
		},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodOptions},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	e.GET("/", handleHome)
	e.GET("/health", handleHealthCheck)
	e.GET("/apps", handlers.HandleListApps(terminalConfig))
	e.GET("/ws", handlers.HandleWebSocket(terminalConfig))
	e.POST("/api/contact", handlers.HandleContact)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("Terminal Backend Server")
	fmt.Println("Apps Directory: ", terminalConfig.AppsDirectory)
	fmt.Println("Available Apps: ", handlers.GetAppsList(terminalConfig))
	fmt.Println("Max Concurrent: ", terminalConfig.MaxConcurrent)
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
