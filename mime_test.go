package marcel

import (
	"bytes"
	"net/mail"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMime(t *testing.T) {
	data := strings.NewReader("oh hi I am an attachment")
	moreData := strings.NewReader("oh hi I am another totally different attachment")

	testEmails := []Email{
		Email{
			To:      "to@example.com",
			From:    "from@example.com",
			ReplyTo: "reply_to@example.com",
			Text:    "this is the text part of a test run",
			HTML:    "this <i>is the HTML part of a test</i> run",
			Subject: "attachment test run",
			Attachments: []Attachment{
				Attachment{
					ContentType: "text/plain",
					Data:        data,
					Filename:    "test_data.txt",
				},
				Attachment{
					ContentType: "text/plain",
					Data:        moreData,
					Filename:    "more_test_data.txt",
				},
			},
		},
		Email{
			To:      "to@example.com",
			From:    "from@example.com",
			ReplyTo: "reply_to@example.com",
			Text:    "this is the text part of a test run",
			HTML:    "this <i>is the HTML part of a test</i> run",
			Subject: "alternative test run",
		},
		Email{
			To:      "to@example.com",
			From:    "from@example.com",
			ReplyTo: "reply_to@example.com",
			Text:    "this is the text, and only, part of a test run",
			Subject: "text only test run",
		},
		Email{
			To:      "to@example.com",
			From:    "from@example.com",
			ReplyTo: "reply_to@example.com",
			HTML:    "this is the <i>html</i>, and only, part of a test run",
			Subject: "html only test run",
		},
	}

	for _, email := range testEmails {

		result, err := email.ToMIME()
		assert.Nil(t, err)

		_, err = mail.ReadMessage(bytes.NewBuffer(result))
		assert.Nil(t, err)
	}
}
