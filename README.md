# Marcel

[![PkgGoDev](https://pkg.go.dev/badge/github.com/davidbanham/marcel)](https://pkg.go.dev/github.com/davidbanham/marcel)

Marcel is a tool to generate IETF compliant emails in raw MIME format. I mainly use this for generating emails with attachments and sending them via amazon SES. If that's what you're doing too, you may want [notifications](https://github.com/davidbanham/notifications)

Marcel supports:
* HTML bodies
* Text bodies
* Emails with only an HTML or Text body
* Attachments
* JSON serialisation of emails and attachments

Marcel does not support:
* Inline attachments

Marcel endeavours to pass the [IETF Msglint](https://tools.ietf.org/tools/msglint/) tool with no errors (aside from ReturnPath). If you have a payload that generates errors please file an issue.

## Example

```Go
package main

import (
	"strings"
  "log"

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
```
