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
			Text:    "this is the text part of a test run",
			HTML:    "this <i>is the <b>HTML</b> part of a test</i> run. And it has LINKS <a href=\"https://google.com\">https://google.com</a>",
			Subject: "rich html test run",
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
		Email{
			To:      "to@example.com",
			From:    "from@example.com",
			ReplyTo: "reply_to@example.com",
			Text:    "this is the text part of a test run",
			HTML:    "this is the <b>HTML</b> part of a test run",
			Subject: "Non-ASCII subject - “–“ - test run",
		},
		Email{
			To:      "to@example.com",
			From:    "from@example.com",
			ReplyTo: "reply_to@example.com",
			Text:    "this is the text part of a test run",
			HTML:    "this is the <b>HTML</b> part of a test run",
			Subject: "Long with non-ASCII - “–“ - test This is a seriously long subject line I mean it is just silly what a ridiculous length of string to put in a subject who would do a think like this it is a bloody outrage do you not know that the maximum length of a MIME header is 75 characters and there's all sorts of nonsense we need to do in order to support multiline headers in combination with encoded words so that non-ASCII characters are supported I mean have you even read rfc2047 20 times?",
		},
		Email{
			To:      "to@example.com",
			From:    "from@example.com",
			ReplyTo: "reply_to@example.com",
			Text:    "this is the text part of a test run",
			HTML:    "this is the <b>HTML</b> part of a test run",
			// Weird double-rune emoji
			Subject: "XXXXX in 👩‍🏫 Totally Great Company Form That is Goodly is now Requires Some Stuff To Happen [Place: Goodstuff OK]",
		},
		Email{
			To:      "to@example.com",
			From:    "from@example.com",
			ReplyTo: "reply_to@example.com",
			Text:    "This contains weird non breaking spaces Dear [[",
			HTML:    "This contains weird non breaking spaces Dear [[",
			Subject: "Text encodings are a pain in the bum.",
		},
		Email{
			To:      "to@example.com",
			From:    "from@example.com",
			ReplyTo: "reply_to@example.com",
			Text:    "this is the text part of a cutom header mail",
			HTML:    "this <i>is the HTML part of a custom header</i> run",
			Subject: "custom header",
			Headers: map[string]string{
				"X-Auto-Response-Suppress": "AutoReply, OOF, RN, NRN",
			},
		},
	}

	for _, email := range testEmails {
		result, err := email.ToMIME()
		assert.Nil(t, err)

		msg, err := mail.ReadMessage(bytes.NewBuffer(result))
		assert.Nil(t, err)
		for k, v := range email.Headers {
			val := msg.Header.Get(k)
			assert.Equal(t, v, val)
		}
	}
}
