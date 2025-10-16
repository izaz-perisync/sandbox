package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	smtp "github.com/emersion/go-smtp"
	"github.com/jhillyerd/enmime"
)

// Simple in-memory credential store
var smtpCreds = map[string]string{
	"user@example.com": "password123",
	"test@uplog.com":   "Pass@123",
}

// Backend implements the SMTP server methods.
type Backend struct{}

func (b *Backend) NewSession(conn *smtp.Conn) (smtp.Session, error) {
	fmt.Println("here")

	return &Session{}, nil
}

// Login authenticates users.
// func (b *Backend) Login(state *smtp.ConnectionInfo, username, password string) (smtp.Session, error) {
// 	if pwd, ok := smtpCreds[username]; !ok || pwd != password {
// 		return nil, fmt.Errorf("invalid username or password")
// 	}
// 	log.Printf("✅ Login success: %s", username)
// 	return &Session{Username: username}, nil
// }

func (s *Session) AuthPlain(username, password string) error {
	fmt.Println("lod here",username)
	if pwd, ok := smtpCreds[username]; ok && pwd == password {
		s.authenticated = true
		s.username = username
		log.Printf("✅ Auth success: %s", username)
		return nil
	}
	log.Printf("❌ Auth failed for: %s", username)
	return fmt.Errorf("invalid username or password")
}
// AuthPlain is called when client uses AUTH PLAIN


// For AUTH LOGIN, we need to handle the challenge-response
func (s *Session) AuthLogin(username, password string) error {
	if pwd, ok := smtpCreds[username]; !ok || pwd != password {
		return fmt.Errorf("invalid username or password")
	}
	s.Username = username
	// s.authed = true
	log.Printf("✅ SMTP AUTH LOGIN success: %s", username)
	return nil
}



// Session stores mail transaction data.
type Session struct {
	Username      string
	from          string
	to            []string
	data          bytes.Buffer
	authenticated bool
	username      string
}

// func (s *Session) AuthPlain(username, password string) error {
// 	if pwd, ok := smtpCreds[username]; ok && pwd == password {
// 		s.authenticated = true
// 		s.username = username
// 		log.Printf("✅ Auth success: %s", username)
// 		return nil
// 	}
// 	return fmt.Errorf("invalid username or password")
// }

// Auth called by Client.Auth()
// func (s *Session) Auth(a smtp.AuthSession) error {
// 	if a == nil {
// 		return nil
// 	}

// 	fmt.Printf("%+v\n", a)

// 	a.AuthMechanisms()
// 	// For simplicity, check username/password in SASL PLAIN payload
// 	// if a.Username() == "user@example.com" && a.Password() == "password123" {
// 	// 	s.authenticated = true
// 	// 	s.username = a.Username()
// 	// 	log.Printf("✅ Auth success: %s", s.username)
// 	// 	return nil
// 	// }
// 	return smtp.ErrAuthUnsupported
// }

// Mail sets the envelope sender.
func (s *Session) Mail(from string, opts *smtp.MailOptions) error {
	s.from = from
	log.Printf("MAIL FROM: %s", from)
	return nil
}

// Rcpt adds a recipient.
func (s *Session) Rcpt(to string, opts *smtp.RcptOptions) error {
	s.to = append(s.to, to)
	log.Printf("RCPT TO: %s", to)
	return nil
}

// Data receives the email content.
func (s *Session) Data(r io.Reader) error {
	s.data.Reset()
	if _, err := io.Copy(&s.data, r); err != nil {
		return err
	}

	// Parse MIME message
	env, err := enmime.ReadEnvelope(bytes.NewReader(s.data.Bytes()))
	if err != nil {
		log.Printf("⚠️ Failed to parse MIME: %v", err)
	}

	// Extract attachments
	type Attachment struct {
		Filename    string `json:"filename"`
		ContentType string `json:"content_type"`
		Size        int    `json:"size"`
		DataBase64  string `json:"data_base64,omitempty"`
	}

	var attachments []Attachment
	if env != nil {
		for _, a := range env.Attachments {
			attachments = append(attachments, Attachment{
				Filename:    a.FileName,
				ContentType: a.ContentType,
				Size:        len(a.Content),
				DataBase64:  base64.StdEncoding.EncodeToString(a.Content),
			})
		}
	}

	// Prepare log payload
	payload := map[string]any{
		"user":        s.Username,
		"from":        s.from,
		"to":          s.to,
		"subject":     env.GetHeader("Subject"),
		"text_body":   env.Text,
		"html_body":   env.HTML,
		"attachments": attachments,
		"received_at": time.Now().Format(time.RFC3339),
	}

	j, _ := json.MarshalIndent(payload, "", "  ")
	log.Println("📩 Received message:\n" + string(j))
	return nil
}

// Reset clears the session data.
func (s *Session) Reset() {
	s.from = ""
	s.to = nil
	s.data.Reset()
}

// Logout cleans up after user disconnects.
func (s *Session) Logout() error {
	log.Printf("👋 User %s logged out", s.Username)
	return nil
}

func main() {
	be := &Backend{}

	server := smtp.NewServer(be)
	fmt.Printf("%+v\n", server)
	server.Addr = ":2525"
	server.Domain = "localhost"
	server.AllowInsecureAuth = true // only for local dev/testing
	server.ReadTimeout = 10 * time.Second
	server.WriteTimeout = 10 * time.Second
	server.MaxMessageBytes = 50 * 1024 * 1024 // 50MB

	server.AllowInsecureAuth = true // for local testing

	// Advertise PLAIN auth so clients like net/smtp can authenticate
	// server.AuthDisabled = false
	// server.AuthMechanisms = []string{"PLAIN"}
	log.Printf("🚀 Mock SMTP server started on %s", server.Addr)
	
	// log.Printf("   Test credentials: user@example.com / password123")

	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	if err := server.Serve(listener); err != nil {
		log.Fatalf("SMTP server error: %v", err)
	}
}

// package main

// import (
// 	"bytes"
// 	"encoding/base64"
// 	"encoding/json"
// 	"io"
// 	"log"
// 	"net"
// 	"time"

// 	"github.com/emersion/go-smtp"
// 	"github.com/jhillyerd/enmime"
// )

// // In-memory credentials
// var smtpCreds = map[string]string{
// 	"user@example.com": "password123",
// }

// // Backend implements the go-smtp Backend interface
// type Backend struct{}

// func (b *Backend) NewSession(c *smtp.Conn) (smtp.Session, error) {
// 	// log.Println("📡 New SMTP connection from:", c.RemoteAddr())
// 	return &Session{}, nil
// }

// // Session implements smtp.Session
// type Session struct {
// 	authenticated bool
// 	username      string
// 	from          string
// 	to            []string
// 	data          bytes.Buffer
// }

// // AuthPlain will be called by clients like swaks
// func (s *Session) AuthPlain(username, password string) error {
// 	if pwd, ok := smtpCreds[username]; ok && pwd == password {
// 		s.authenticated = true
// 		s.username = username
// 		log.Printf("✅ Auth success: %s", username)
// 		return nil
// 	}
// 	log.Printf("❌ Auth failed for: %s", username)
// 	return smtp.ErrAuthUnsupported
// }

// func (s *Session) Mail(from string, opts *smtp.MailOptions) error {
// 	if !s.authenticated {
// 		return smtp.ErrAuthRequired
// 	}
// 	s.from = from
// 	return nil
// }

// func (s *Session) Rcpt(to string, opts *smtp.RcptOptions) error {
// 	s.to = append(s.to, to)
// 	return nil
// }

// func (s *Session) Data(r io.Reader) error {
// 	s.data.Reset()
// 	if _, err := io.Copy(&s.data, r); err != nil {
// 		return err
// 	}

// 	env, _ := enmime.ReadEnvelope(bytes.NewReader(s.data.Bytes()))
// 	attachments := []map[string]any{}
// 	if env != nil {
// 		for _, a := range env.Attachments {
// 			attachments = append(attachments, map[string]any{
// 				"filename":    a.FileName,
// 				"contentType": a.ContentType,
// 				"size":        len(a.Content),
// 				"data_b64":    base64.StdEncoding.EncodeToString(a.Content),
// 			})
// 		}
// 	}

// 	payload := map[string]any{
// 		"user":        s.username,
// 		"from":        s.from,
// 		"to":          s.to,
// 		"subject":     env.GetHeader("Subject"),
// 		"text_body":   env.Text,
// 		"html_body":   env.HTML,
// 		"attachments": attachments,
// 		"received_at": time.Now().Format(time.RFC3339),
// 	}

// 	j, _ := json.MarshalIndent(payload, "", "  ")
// 	log.Println("📩 Received email:\n" + string(j))
// 	return nil
// }

// func (s *Session) Reset() {
// 	s.from = ""
// 	s.to = nil
// 	s.data.Reset()
// }

// func (s *Session) Logout() error {
// 	return nil
// }

// func main() {
// 	be := &Backend{}
// 	server := smtp.NewServer(be)

// 	server.Addr = ":2525"
// 	server.Domain = "localhost"
// 	server.AllowInsecureAuth = true
// 	server.ReadTimeout = 10 * time.Second
// 	server.WriteTimeout = 10 * time.Second
// 	server.MaxMessageBytes = 50 * 1024 * 1024

// 	log.Println("🚀 Mock SMTP server running on :2525 (AuthPlain supported)")

// 	ln, err := net.Listen("tcp", server.Addr)
// 	if err != nil {
// 		log.Fatalf("failed to listen: %v", err)
// 	}
// 	if err := server.Serve(ln); err != nil {
// 		log.Fatalf("SMTP server error: %v", err)
// 	}
// }
