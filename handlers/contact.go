// Package handlers contains the funcs for the routes
package handlers

import (
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"regexp"
	"strings"

	"github.com/labstack/echo/v4"
)

type ContactRequest struct {
	Name    string `json:"name" validate:"required"`
	Email   string `json:"email" validate:"required,email"`
	Subject string `json:"subject"`
	Message string `json:"message" validate:"required"`
}

type ContactResponse struct {
	Success bool   `json:"success,omitempty"`
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

func HandleContact(c echo.Context) error {
	var req ContactRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ContactResponse{
			Error: "Invalid request Body",
		})
	}

	if req.Name == "" || req.Email == "" || req.Message == "" {
		return c.JSON(http.StatusBadRequest, ContactResponse{
			Error: "Missing Required Fields",
		})
	}

	if !isValidEmail(req.Email) {
		return c.JSON(http.StatusBadRequest, ContactResponse{
			Error: "Invalid email address",
		})
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.TrimSpace(req.Email)
	req.Subject = strings.TrimSpace(req.Subject)
	req.Message = strings.TrimSpace(req.Message)

	if err := sendEmail(req); err != nil {
		log.Printf("Error sendign email: %v", err)
		return c.JSON(http.StatusInternalServerError, ContactResponse{
			Error: "Failed to send message",
		})
	}

	return c.JSON(http.StatusOK, ContactResponse{
		Success: true,
		Message: "Message sent successfully!",
	})
}

func sendEmail(req ContactRequest) error {
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"
	from := os.Getenv("GMAIL_FROM")
	password := os.Getenv("GMAIL_PASSWORD")
	to := os.Getenv("RECIPIENT_EMAIL")

	subject := "New Contact Form Submission"
	if req.Subject != "" {
		subject = fmt.Sprintf("Contact Form: %s", req.Subject)
	}

	body := fmt.Sprintf(
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/plain; charset=UTF-8\r\n"+
			"\r\n"+
			"You received a new message from your contact form:\r\n\r\n"+
			"Name: %s\r\n"+
			"Email: %s\r\n"+
			"Subject: %s\r\n\r\n"+
			"Message:\r\n%s\r\n",
		from,
		to,
		subject,
		req.Name,
		req.Email,
		req.Subject,
		req.Message,
	)

	auth := smtp.PlainAuth("", from, password, smtpHost)

	addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)
	err := smtp.SendMail(addr, auth, from, []string{to}, []byte(body))

	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}
