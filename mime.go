package marcel

import (
	"bytes"
	"encoding/base64"
	"io"
	"mime/multipart"
	"net/textproto"
	"time"
)

type Email struct {
	To          string
	From        string
	ReplyTo     string
	Text        string
	HTML        string
	Subject     string
	Attachments []Attachment
}

type Attachment struct {
	ContentType string
	Data        io.Reader
	Filename    string
}

//  multipart/mixed
//  |- multipart/related
//  |  |- multipart/alternative
//  |  |  |- text/plain
//  |  |  `- text/html
//  |  `- inlines..
//  `- attachments..

func (email Email) ToMIME() ([]byte, error) {
	mixedContent := &bytes.Buffer{}
	mixedWriter := multipart.NewWriter(mixedContent)

	// related content, inside mixed
	var newBoundary = "RELATED-" + mixedWriter.Boundary()
	mixedWriter.SetBoundary(first70("MIXED-" + mixedWriter.Boundary()))

	relatedWriter, newBoundary, err := nestedMultipart(mixedWriter, "multipart/related", newBoundary)
	if err != nil {
		return []byte{}, err
	}
	altWriter, newBoundary, err := nestedMultipart(relatedWriter, "multipart/alternative", "ALTERNATIVE-"+newBoundary)
	if err != nil {
		return []byte{}, err
	}

	if email.Text != "" {
		childContent, err := altWriter.CreatePart(textproto.MIMEHeader{"Content-Type": {"text/plain"}})
		if err != nil {
			return []byte{}, err
		}
		childContent.Write([]byte(email.Text + "\r\n\r\n"))
	}
	if email.HTML != "" {
		childContent, err := altWriter.CreatePart(textproto.MIMEHeader{"Content-Type": {"text/html"}})
		if err != nil {
			return []byte{}, err
		}
		childContent.Write([]byte(email.HTML + "\r\n"))
	}

	if err := altWriter.Close(); err != nil {
		return []byte{}, err
	}
	if err := relatedWriter.Close(); err != nil {
		return []byte{}, err
	}

	// Attachments
	for _, attachment := range email.Attachments {
		fileContent, err := mixedWriter.CreatePart(textproto.MIMEHeader{
			"Content-Type": {
				attachment.ContentType + "; name=\"" + attachment.Filename + "\"",
			}, "Content-Disposition": {
				"attachment; filename=\"" + attachment.Filename + "\"",
			}, "Content-Transfer-Encoding": {
				"base64",
			},
		})
		if err != nil {
			return []byte{}, err
		}
		//func NewEncoder(enc *Encoding, w io.Writer) io.WriteCloser {
		enc := base64.NewEncoder(base64.StdEncoding, fileContent)
		if _, err := io.Copy(enc, attachment.Data); err != nil {
			return []byte{}, err
		}
		if err := enc.Close(); err != nil {
			return []byte{}, err
		}
		fileContent.Write([]byte("\r\n\r\n"))
	}

	if err := mixedWriter.Close(); err != nil {
		return []byte{}, err
	}

	rawEmail := []byte{}
	rawEmail = append(rawEmail, []byte("From: "+email.From+"\r\n")...)
	rawEmail = append(rawEmail, []byte("MIME-Version: 1.0\r\n")...)
	rawEmail = append(rawEmail, []byte("Date: "+time.Now().Format(time.RFC1123Z)+"\r\n")...)
	rawEmail = append(rawEmail, []byte("To: "+email.To+"\r\n")...)
	rawEmail = append(rawEmail, []byte("Subject: "+email.Subject+"\r\n")...)
	rawEmail = append(rawEmail, []byte("Content-Type: multipart/mixed; boundary=")...)
	rawEmail = append(rawEmail, []byte(`"`+mixedWriter.Boundary()+"\"\r\n\r\n")...)
	rawEmail = append(rawEmail, []byte(mixedContent.String())...)

	return rawEmail, nil

}

func nestedMultipart(enclosingWriter *multipart.Writer, contentType, boundary string) (*multipart.Writer, string, error) {
	var nestedWriter *multipart.Writer
	var newBoundary string

	boundary = first70(boundary)
	contentWithBoundary := contentType + "; boundary=\"" + boundary + "\""
	if contentType == "multipart/related" {
		contentWithBoundary += "; type=\"Text/HTML\""
	}
	contentBuffer, err := enclosingWriter.CreatePart(textproto.MIMEHeader{"Content-Type": {contentWithBoundary}})
	if err != nil {
		return nestedWriter, newBoundary, err
	}

	nestedWriter = multipart.NewWriter(contentBuffer)
	newBoundary = nestedWriter.Boundary()
	nestedWriter.SetBoundary(boundary)
	return nestedWriter, newBoundary, nil
}

func first70(str string) string {
	if len(str) > 70 {
		return string(str[0:69])
	}
	return str
}
