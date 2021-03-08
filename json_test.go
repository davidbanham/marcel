package marcel

import (
	"encoding/json"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSONMarshalAndUnmarshal(t *testing.T) {
	data := strings.NewReader("oh hi I am an attachment")

	email := Email{
		To:      "to@example.com",
		From:    "from@example.com",
		ReplyTo: "reply_to@example.com",
		Text:    "this is the text part of a test run",
		HTML:    "this <i>is the HTML part of a test</i> run",
		Subject: "test run",
		Attachments: []Attachment{
			Attachment{
				ContentType: "text/plain",
				Data:        data,
				Filename:    "test_data.txt",
			},
		},
	}

	result, err := json.Marshal(email)
	assert.Nil(t, err)

	newEmail := Email{}

	assert.Nil(t, json.Unmarshal(result, &newEmail))

	contents, err := ioutil.ReadAll(newEmail.Attachments[0].Data)
	assert.Nil(t, err)
	assert.Equal(t, string(contents), "oh hi I am an attachment")
}
