package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/davidbanham/marcel"
)

func main() {
	mail := marcel.Email{
		To:      "to@example.com",
		From:    "from@example.com",
		ReplyTo: "reply_to@example.com", // optional
		Text:    "This is the important information in text format",
		HTML:    "This is the important information in <b>HTML</b> format",
		Subject: "A really important email",
		Attachments: []marcel.Attachment{
			marcel.Attachment{
				ContentType: "text/plain",
				Data:        strings.NewReader("A very important attachment"), // Data will be base64 encoded before sending
				Filename:    "test_data.txt",
			},
		},
	}

	rawEmail, err := mail.ToMIME()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(rawEmail))
}
