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
	"fmt"
	"log"
	"net/smtp"
	"strings"
)

func main() {
	from := "test@sender.com"
	to := []string{"user1@example.com", "user2@example.com"}
	cc := []string{"team1@example.com", "team2@example.com"}
	bcc := []string{"hidden1@example.com", "hidden2@example.com"}
	replyTo := []string{"support@sender.com", "no-reply@sender.com"}
	

	subject := "Test Email with CC and BCC"
	htmlBody := `
    <html>
    <body>
        <p>Dear User,</p>
        <p>This is a test email with CC and BCC recipients.</p>
    </body>
    </html>`

	// Build message headers only
	var msg bytes.Buffer

	// Headers
	msg.WriteString(fmt.Sprintf("From: %s\r\n", from))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(to, ", ")))

	if len(cc) > 0 {
		msg.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(cc, ", ")))
	}

	if len(bcc)>0{
		msg.WriteString(fmt.Sprintf("Bcc: %s\r\n", strings.Join(bcc, ", ")))
	}

	if len(replyTo) > 0 {
		msg.WriteString(fmt.Sprintf("Reply-To: %s\r\n", strings.Join(replyTo, ", ")))
	}
	msg.WriteString(fmt.Sprintf("Return-Path:%s\r\n","dev2335"))

	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: text/html; charset=\"utf-8\"\r\n")
	msg.WriteString("\r\n") // Empty line separates headers from body

	// Body
	msg.WriteString(htmlBody + "\r\n")

	// Combine ALL recipients (TO + CC + BCC) for RCPT TO commands
	allRecipients := append(append([]string{}, to...), cc...)
	allRecipients = append(allRecipients, bcc...)

	fmt.Printf("Sending to %d recipients:\n", len(allRecipients))
	fmt.Printf("  TO: %v\n", to)
	fmt.Printf("  CC: %v\n", cc)
	fmt.Printf("  BCC: %v\n", bcc)

	// Send email
	auth := smtp.PlainAuth("", "user_0Ku1Qry7", "pass_RIeKnB6c+GY=", "localhost")
	if err := smtp.SendMail("localhost:8081", auth, from, allRecipients, msg.Bytes()); err != nil {
		log.Fatal("SendMail failed:", err)
	}

	log.Println("✅ Email with CC/BCC sent successfully")
}

// func main() {
// 	from := "test@sender.com"
// 	// to := "user@example.com"
// 	to := []string{"user1@example.com", "user2@example.com"}
// 	cc := []string{"team1@example.com", "team2@example.com"}
// 	bcc := []string{"hidden1@example.com", "hidden2@example.com"}
// 	replyTo := []string{"support@sender.com", "no-reply@sender.com"}
// 	// Boundary for separating parts
// 	boundary := "MYBOUNDARY123"

// 	subject := "Inline Image Email"
// 	htmlBody := `
// 	<html>
// <head><title>ignore me</title></head>
// <body>
//   <p>Dear alone dhanu,</p>
//   <p>A new release of Calendria is now available for testing.</p>
//   <p>Or open the <b>AppShare mobile app</b> and tap the release to install.</p>
//   <hr>
//   <p>Need Help?</p>
//   <p>Contact Us at <a href="https://zunoy.com/contact-us">Zunoy</a></p>
// </body>
// </html>`

// 	// Optional: clean base64 and validate image
// 	htmlBody = sanitizeBase64Images(htmlBody)

// 	// filename := "hello.txt"
// 	// fileData := []byte("This is a small file content example.")
// 	// encodedFile := base64.StdEncoding.EncodeToString(fileData)
// 	filePath := "sample.png" // put your file path here
// 	imageData, err := os.ReadFile(filePath)
// 	if err != nil {
// 		log.Fatalf("failed to read image file: %v", err)
// 	}

// 	filename := filepath.Base(filePath)
// 	encodedFile := base64.StdEncoding.EncodeToString(imageData)
// 	// Build message
// 	var msg bytes.Buffer
// 	msg.WriteString(fmt.Sprintf("From: %s\r\n", from))
// 	msg.WriteString(fmt.Sprintf("To: %s\r\n", to))
// 	msg.WriteString("Subject: Test Email with HTML and Attachment\r\n")
// 	msg.WriteString("MIME-Version: 1.0\r\n")
// 	// msg.WriteString("Content-Type: multipart/mixed; boundary=" + boundary + "\r\n")
// 	msg.WriteString(fmt.Sprintf("Content-Type: multipart/related; boundary=%s\r\n", boundary))
// 	msg.WriteString("\r\n")

// 	// HTML Part
// 	msg.WriteString("--" + boundary + "\r\n")
// 	msg.WriteString("Content-Type: text/html; charset=\"utf-8\"\r\n")
// 	msg.WriteString("Content-Transfer-Encoding: 7bit\r\n\r\n")
// 	msg.WriteString("\r\n")
// 	msg.WriteString(htmlBody + "\r\n")

// 	// Attachment Part
// 	msg.WriteString("--" + boundary + "\r\n")
// 	msg.WriteString("Content-Type: text/plain; name=\"" + filename + "\"\r\n")
// 	msg.WriteString("Content-Disposition: attachment; filename=\"" + filename + "\"\r\n")
// 	msg.WriteString("Content-Transfer-Encoding: base64\r\n")
// 	msg.WriteString("Content-ID: <image1>\r\n\r\n")
// 	msg.WriteString("\r\n")

// 	// Build MIME message

// 	msg.WriteString(fmt.Sprintf("From: %s\r\n", from))
// 	msg.WriteString(fmt.Sprintf("To: %s\r\n", to))
// 	msg.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(to, ", ")))
// 	msg.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(cc, ", ")))
// 	msg.WriteString(fmt.Sprintf("Reply-To: %s\r\n", strings.Join(replyTo, ", ")))
// 	msg.WriteString(fmt.Sprintf("Bcc: %s\r\n", strings.Join(bcc, ", ")))
// 	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))

// 	msg.WriteString("MIME-Version: 1.0\r\n")
// 	msg.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=%s\r\n", boundary))
// 	msg.WriteString("\r\n")

// 	// HTML part
// 	msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
// 	msg.WriteString("Content-Type: text/html; charset=\"utf-8\"\r\n")
// 	msg.WriteString("Content-Transfer-Encoding: quoted-printable\r\n\r\n")
// 	msg.WriteString(htmlBody + "\r\n")

// 	msg.WriteString(fmt.Sprintf("--%s--\r\n", boundary))

// 	for i := 0; i < len(encodedFile); i += 76 {
// 		end := i + 76
// 		if end > len(encodedFile) {
// 			end = len(encodedFile)
// 		}
// 		msg.WriteString(encodedFile[i:end] + "\r\n")
// 	}
// 	msg.WriteString("--" + boundary + "--\r\n")

// 	// Send
// 	auth := smtp.PlainAuth("", "user_0Ku1Qry7", "pass_RIeKnB6c+GY=", "localhost")
// 	if err := smtp.SendMail("localhost:8081", auth, from, to, msg.Bytes()); err != nil {
// 		log.Fatal("SendMail failed:", err)
// 	}

// 	// auth := smtp.PlainAuth("", "user_0Ku1Qry7", "pass_RIeKnB6c+GY=", "localhost")

// 	// if err := smtp.SendMail("localhost:8081", auth, from, to, []byte("Subject: test\r\n\r\nHello!"));err != nil {
// 	// 	log.Fatal("SendMail failed:", err)
// 	// }
// 	log.Println("✅ Email with inline image sent successfully")
// }

// sanitizeBase64Images ensures only valid data URIs are kept, strips long junk
func sanitizeBase64Images(html string) string {
	return strings.ReplaceAll(html, "\n", "")
}
