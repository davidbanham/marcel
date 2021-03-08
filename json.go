package marcel

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
)

func (attachment *Attachment) MarshalJSON() ([]byte, error) {
	data, err := ioutil.ReadAll(attachment.Data)
	if err != nil {
		return []byte{}, err
	}

	return json.Marshal(&struct {
		ContentType string `json:"content_type"`
		Data        []byte `json:"data"`
		Filename    string `json:"filename"`
	}{
		ContentType: attachment.ContentType,
		Data:        data,
		Filename:    attachment.Filename,
	})
}

func (attachment *Attachment) UnmarshalJSON(data []byte) error {
	inner := struct {
		ContentType string `json:"content_type"`
		Data        []byte `json:"data"`
		Filename    string `json:"filename"`
	}{}

	if err := json.Unmarshal(data, &inner); err != nil {
		return err
	}

	attachment.ContentType = inner.ContentType
	attachment.Filename = inner.Filename
	attachment.Data = bytes.NewReader(inner.Data)

	return nil
}
