package services

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
)

func SendEmail(email, textToBeSent string) {
	from := os.Getenv("EMAIL_FROM")
    pass := os.Getenv("EMAIL_PASSWORD")
    to := email

    msg := "From: " + from + "\n" +
        "To: " + to + "\n" +
        "Subject: Hello there\n\n" +
        textToBeSent

    err := smtp.SendMail(fmt.Sprintf("%s:%s", os.Getenv("EMAIL_HOST"), os.Getenv("EMAIL_PORT")),
        smtp.PlainAuth("", from, pass, os.Getenv("EMAIL_HOST")),
        from, []string{to}, []byte(msg))

    if err != nil {
        log.Printf("smtp error: %s", err)
        return
    }
    log.Println("Successfully sended to " + to)
}