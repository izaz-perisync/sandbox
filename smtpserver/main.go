package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net"
	"net/textproto"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// In-memory username/password map
var smtpCreds = map[string]string{
	"user@example.com": "password123",
}

// Simple email struct for logging
type Email struct {
	User       string       `json:"user"`
	From       string       `json:"from"`
	To         []string     `json:"to"`
	Subject    string       `json:"subject"`
	Body       string       `json:"body"`
	Text       string       `json:"text"`
	Attachment []Attachment `json:"Attachment"`
	ReceivedAt string       `json:"received_at"`
}

type Attachment struct {
	PartID      string `json:"PartID"`
	FileName    string `json:"FileName"`
	ContentType string `json:"ContentType"`
	ContentID   string `json:"ContentID"`
	Size        int    `json:"Size"`
}

func handleConn(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	send := func(msg string) {
		writer.WriteString(msg + "\r\n")
		writer.Flush()
	}

	send("220 localhost Mock SMTP Server Ready")

	authenticated := false
	var username string
	var from string
	var to []string
	inData := false
	var data []string
	authType := ""

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		line = strings.TrimSpace(line)
		// log.Println("C:", line)
		if inData {
			if line == "." {
				inData = false
				send("250 OK: queued")

				fullMessage := strings.Join(data, "\r\n")

				// Parse headers
				reader := textproto.NewReader(bufio.NewReader(strings.NewReader(fullMessage)))
				header, err := reader.ReadMIMEHeader()
				if err != nil {
					log.Println("❌ Failed to read MIME headers:", err)
				}

				subject := header.Get("Subject")
				contentType := header.Get("Content-Type")

				attachments := []Attachment{}
				bodyText := ""

				// Check for multipart email
				mediaType, params, err := mime.ParseMediaType(contentType)
				if err == nil && strings.HasPrefix(mediaType, "multipart/") {
					boundary := params["boundary"]
					mr := multipart.NewReader(strings.NewReader(fullMessage), boundary)

					for {
						p, err := mr.NextPart()
						if err != nil {
							break
						}

						slurp, _ := io.ReadAll(p)
						disposition := p.Header.Get("Content-Disposition")
						contentType := p.Header.Get("Content-Type")

						// Check if it’s an attachment
						if strings.Contains(strings.ToLower(disposition), "attachment") ||
							strings.Contains(strings.ToLower(disposition), "inline") && strings.HasPrefix(contentType, "image/") {
							filename := p.FileName()
							if filename == "" {
								filename = "unnamed"
							}
							// attachments = append(attachments, filename)

							contentID := p.Header.Get("Content-ID")
							// partID := fmt.Sprintf("%d", i) // or however you track part count
							size := len(slurp)

							attachments = append(attachments, Attachment{
								PartID:      "2",
								FileName:    filename,
								ContentType: contentType,
								ContentID:   strings.Trim(contentID, "<>"),
								Size:        size,
							})

							log.Printf("📎 Attachment found: %s (%d bytes, type=%s, cid=%s)\n",
								filename, size, contentType, contentID)

							// Optional: base64 encode or save
							b64 := base64.StdEncoding.EncodeToString(slurp)
							log.Printf("📄 Base64 (first 100 chars): %s\n", b64[:100])

						} else {
							bodyText += string(slurp)
						}
					}
				} else {
					bodyText = fullMessage
				}
				fmt.Println("bodyText", bodyText)
				cleanBody := extractBodyHTML(bodyText)
				fmt.Println("cl", cleanBody)
				fmt.Println("dddd")
				email := Email{
					User:       username,
					From:       from,
					To:         to,
					Subject:    subject,
					Body:       bodyText,
					Attachment: attachments,
					ReceivedAt: time.Now().Format(time.RFC3339),
					Text:       cleanBody,
				}

				j, _ := json.MarshalIndent(email, "", "  ")
				log.Println("📩 Received email:\n" + string(j))

				// Reset for next email
				data = nil
				from = ""
				to = nil
				continue
			}
			data = append(data, line)
			continue
		}

		switch {
		case strings.HasPrefix(line, "EHLO") || strings.HasPrefix(line, "HELO"):
			send("250-localhost")
			send("250-AUTH PLAIN LOGIN")
			send("250 OK")
		case strings.HasPrefix(line, "AUTH PLAIN"):
			// Do NOT treat it as LOGIN
			authType = "PLAIN"
			// fmt.Println("l", line)
			line = strings.TrimSpace(line)
			parts := strings.SplitN(line, " ", 3)

			if len(parts) < 2 {
				send("501 Syntax error")
				return
			}

			b64payload := ""
			if len(parts) == 3 {
				// Base64 payload on same line
				b64payload = strings.TrimSpace(parts[2])
			} else if len(parts) == 2 {
				// Client may send payload next
				send("334 ") // prompt for payload
				b64payload, _ = reader.ReadString('\n')
				b64payload = strings.TrimSpace(b64payload)
			}

			// b64payload := strings.TrimSpace(parts[1])
			decoded, err := base64.StdEncoding.DecodeString(b64payload)
			if err != nil {
				log.Println("base64 decode error:", err, "payload:", b64payload)
				send("501 Invalid base64")
				return
			}

			// PLAIN format: [authzid]\x00username\x00password
			tokens := bytes.SplitN(decoded, []byte{0}, 3)
			if len(tokens) < 2 {
				send("501 Invalid auth format")
				return
			}

			username := string(tokens[len(tokens)-2])
			password := string(tokens[len(tokens)-1])
			log.Println("decoded username:", username, "password:", password)

			if pwd, ok := smtpCreds[username]; ok && pwd == password {
				authenticated = true
				send("235 Authentication successful")
				log.Println("✅ Auth PLAIN success for:", username)
			} else {
				send("535 Authentication failed")
				log.Println("❌ Auth PLAIN failed for:", username)
				return
			}

		case authType == "LOGIN" && !authenticated && username == "":
			decoded, _ := base64.StdEncoding.DecodeString(line)
			username = string(decoded)
			fmt.Println("userName", username)
			send("334 UGFzc3dvcmQ6") // Password:
		case authType == "LOGIN" && !authenticated:
			decoded, _ := base64.StdEncoding.DecodeString(line)
			password := string(decoded)
			fmt.Println("password", password)
			if pwd, ok := smtpCreds[username]; ok && pwd == password {
				authenticated = true
				send("235 Authentication successful")
				log.Println("✅ Auth LOGIN success for:", username)
			} else {
				send("535 Authentication failed")
				log.Println("❌ Auth LOGIN failed for:", username)
				return
			}
		case strings.HasPrefix(line, "MAIL FROM:"):
			from = strings.Trim(line[len("MAIL FROM:"):], "<>")
			send("250 OK")
		case strings.HasPrefix(line, "RCPT TO:"):
			to = append(to, strings.Trim(line[len("RCPT TO:"):], "<>"))
			send("250 OK")
		case strings.HasPrefix(line, "DATA"):
			send("354 End data with <CR><LF>.<CR><LF>")
			inData = true
		case strings.HasPrefix(line, "QUIT"):
			send("221 Bye")
			return
		default:
			send("250 OK")
		}
	}
}

func main() {
	ln, err := net.Listen("tcp", ":2525")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("🚀 Mock SMTP server running on :2525")

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Accept error:", err)
			continue
		}
		go handleConn(conn)
	}
}

func decodeHTML(encoded string) string {
	replacer := strings.NewReplacer(
		"\\u003c", "<",
		"\\u003e", ">",
		"\\u0026", "&",
		"\\u003d", "=",
		"\\u0022", `"`,
	)
	return replacer.Replace(encoded)
}

func extractBodyHTML(htmlStr string) string {
	htmlStr = decodeHTML(htmlStr)

	doc, err := html.Parse(strings.NewReader(htmlStr))
	if err != nil {
		log.Println("parse err:", err)
		return htmlStr
	}

	var sb strings.Builder
	var inBody bool

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "body" {
			inBody = true
		}
		if inBody {
			switch n.Type {
			case html.TextNode:
				text := strings.TrimSpace(n.Data)
				if text != "" {
					sb.WriteString(text)
					sb.WriteString(" ")
				}
			case html.ElementNode:
				switch n.Data {
				case "br":
					sb.WriteString("\n")
				case "p", "div", "section", "li", "tr", "td", "h1", "h2", "h3":
					sb.WriteString("\n\n")
				case "hr":
					sb.WriteString("\n----------\n")
				case "a":
					for _, attr := range n.Attr {
						if attr.Key == "href" {
							sb.WriteString(" (" + attr.Val + ") ")
						}
					}
				case "img":
					var alt, src string
					for _, attr := range n.Attr {
						if attr.Key == "alt" {
							alt = attr.Val
						}
						if attr.Key == "src" {
							src = attr.Val
						}
					}
					if alt != "" || src != "" {
						sb.WriteString(fmt.Sprintf("%s (%s) ", alt, src))
					}
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}

		if n.Type == html.ElementNode && n.Data == "body" {
			inBody = false
		}
	}

	walk(doc)
	return strings.TrimSpace(sb.String())
}
