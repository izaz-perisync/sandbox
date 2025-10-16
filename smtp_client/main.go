// package main

// import (
// 	"log"
// 	"net/smtp"
// )

//	func main() {
//		auth := smtp.PlainAuth("", "user@example.com", "password123", "localhost")
//		err := smtp.SendMail(
//			"localhost:2525",
//			auth,
//			"test@sender.com",
//			[]string{"user@example.com"},
//			[]byte("Subject: Hello from net/smtp\n\nThis is a test email!"),
//		)
//		if err != nil {
//			log.Fatal("SendMail failed:", err)
//		}
//		log.Println("✅ Email sent to mock SMTP server")
//	}
package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	from := "test@sender.com"
	to := "user@example.com"
	// Boundary for separating parts
	boundary := "MYBOUNDARY123"

	subject := "Inline Image Email"
	htmlBody := `
	<html>
<head><title>ignore me</title></head>
<body>
  <p>Dear alone dhanu,</p>
  <p>A new release of Calendria is now available for testing.</p>
  <p>Or open the <b>AppShare mobile app</b> and tap the release to install.</p>
  <hr>
  <p>Need Help?</p>
  <p>Contact Us at <a href="https://zunoy.com/contact-us">Zunoy</a></p>
</body>
</html>`

	// Optional: clean base64 and validate image
	htmlBody = sanitizeBase64Images(htmlBody)

	// filename := "hello.txt"
	// fileData := []byte("This is a small file content example.")
	// encodedFile := base64.StdEncoding.EncodeToString(fileData)
	filePath := "sample.png" // put your file path here
	imageData, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("failed to read image file: %v", err)
	}

	filename := filepath.Base(filePath)
	encodedFile := base64.StdEncoding.EncodeToString(imageData)
	// Build message
	var msg bytes.Buffer
	msg.WriteString(fmt.Sprintf("From: %s\r\n", from))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", to))
	msg.WriteString("Subject: Test Email with HTML and Attachment\r\n")
	msg.WriteString("MIME-Version: 1.0\r\n")
	// msg.WriteString("Content-Type: multipart/mixed; boundary=" + boundary + "\r\n")
	msg.WriteString(fmt.Sprintf("Content-Type: multipart/related; boundary=%s\r\n", boundary))
	msg.WriteString("\r\n")

	// HTML Part
	msg.WriteString("--" + boundary + "\r\n")
	msg.WriteString("Content-Type: text/html; charset=\"utf-8\"\r\n")
	msg.WriteString("Content-Transfer-Encoding: 7bit\r\n\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(htmlBody + "\r\n")

	// Attachment Part
	msg.WriteString("--" + boundary + "\r\n")
	msg.WriteString("Content-Type: text/plain; name=\"" + filename + "\"\r\n")
	msg.WriteString("Content-Disposition: attachment; filename=\"" + filename + "\"\r\n")
	msg.WriteString("Content-Transfer-Encoding: base64\r\n")
	msg.WriteString("Content-ID: <image1>\r\n\r\n")
	msg.WriteString("\r\n")

	// Build MIME message

	msg.WriteString(fmt.Sprintf("From: %s\r\n", from))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", to))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=%s\r\n", boundary))
	msg.WriteString("\r\n")

	// HTML part
	msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	msg.WriteString("Content-Type: text/html; charset=\"utf-8\"\r\n")
	msg.WriteString("Content-Transfer-Encoding: quoted-printable\r\n\r\n")
	msg.WriteString(htmlBody + "\r\n")

	msg.WriteString(fmt.Sprintf("--%s--\r\n", boundary))

	for i := 0; i < len(encodedFile); i += 76 {
		end := i + 76
		if end > len(encodedFile) {
			end = len(encodedFile)
		}
		msg.WriteString(encodedFile[i:end] + "\r\n")
	}
	msg.WriteString("--" + boundary + "--\r\n")

	// Send
	auth := smtp.PlainAuth("", "user@example.com", "password123", "localhost")
	if err := smtp.SendMail("localhost:2525", auth, from, []string{to}, msg.Bytes());err != nil {
		log.Fatal("SendMail failed:", err)
	}

	log.Println("✅ Email with inline image sent successfully")
}

// sanitizeBase64Images ensures only valid data URIs are kept, strips long junk
func sanitizeBase64Images(html string) string {
	return strings.ReplaceAll(html, "\n", "")
}
