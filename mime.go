package marcel

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"mime/multipart"
	"mime/quotedprintable"
	"net/textproto"
	"strings"
	"time"
	"unicode"
)

type Email struct {
	To          string
	From        string
	ReplyTo     string
	ReturnPath  string
	Text        string
	HTML        string
	Subject     string
	Attachments []Attachment
	Headers     map[string]string
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
	dest := bytes.NewBuffer([]byte{})
	if err := email.WriteMime(dest); err != nil {
		return []byte{}, err
	}
	return io.ReadAll(dest)
}

func (email Email) WriteMime(dest io.Writer) error {
	mixedContent := &bytes.Buffer{}
	mixedWriter := multipart.NewWriter(mixedContent)

	// related content, inside mixed
	var relatedBoundary = "RELATED-" + mixedWriter.Boundary()
	mixedWriter.SetBoundary(first70("MIXED-" + mixedWriter.Boundary()))

	relatedWriter, alternativeBoundary, err := nestedMultipart(mixedWriter, "multipart/related", relatedBoundary)
	if err != nil {
		return err
	}
	altWriter, _, err := nestedMultipart(relatedWriter, "multipart/alternative", "ALTERNATIVE-"+alternativeBoundary)
	if err != nil {
		return err
	}

	if email.Text != "" {
		childContent, err := altWriter.CreatePart(textproto.MIMEHeader{
			"Content-Type":              {"text/plain; charset=\"UTF-8\""},
			"Content-Transfer-Encoding": {"quoted-printable"},
			"Content-Disposition":       {"inline"},
		})
		if err != nil {
			return err
		}
		enc := quotedprintable.NewWriter(childContent)
		enc.Write([]byte(email.Text + "\r\n\r\n"))
	}
	if email.HTML != "" {
		childContent, err := altWriter.CreatePart(textproto.MIMEHeader{
			"Content-Type":              {"text/html; charset=\"UTF-8\""},
			"Content-Transfer-Encoding": {"quoted-printable"},
			"Content-Disposition":       {"inline"},
		})
		if err != nil {
			return err
		}
		enc := quotedprintable.NewWriter(childContent)
		enc.Write([]byte(email.HTML + "\r\n\r\n"))
	}

	if err := altWriter.Close(); err != nil {
		return err
	}
	if err := relatedWriter.Close(); err != nil {
		return err
	}

	for k, v := range email.Headers {
		dest.Write([]byte(fmt.Sprintf("%s: %s", k, v) + "\r\n"))
	}

	dest.Write([]byte("From: " + email.From + "\r\n"))
	dest.Write([]byte("To: " + email.To + "\r\n"))
	if email.ReturnPath != "" {
		if string(email.ReturnPath[0]) == "<" {
			dest.Write([]byte("Return-Path: " + email.ReturnPath + "\r\n"))
		} else {
			dest.Write([]byte("Return-Path: <" + email.ReturnPath + ">\r\n"))
		}
	}
	if email.ReplyTo != "" {
		dest.Write([]byte("Reply-To: " + email.ReplyTo + "\r\n"))
	}

	subjectHeader, err := encodeHeader("Subject", email.Subject)
	if err != nil {
		return err
	}
	dest.Write([]byte(subjectHeader))

	dest.Write([]byte("Date: " + time.Now().Format(time.RFC1123Z) + "\r\n"))
	dest.Write([]byte("MIME-Version: 1.0\r\n"))

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
			return err
		}
		enc := base64.NewEncoder(base64.StdEncoding, fileContent)
		if _, err := io.Copy(enc, attachment.Data); err != nil {
			return err
		}
		if err := enc.Close(); err != nil {
			return err
		}
		fileContent.Write([]byte("\r\n\r\n"))
	}

	dest.Write([]byte("Content-Type: multipart/mixed; boundary="))
	dest.Write([]byte(`"` + mixedWriter.Boundary() + "\"\r\n\r\n"))
	if _, err := io.Copy(dest, mixedContent); err != nil {
		return err
	}
	dest.Write([]byte("--" + mixedWriter.Boundary() + "--\r\n\r\n"))

	if err := mixedWriter.Close(); err != nil {
		return err
	}

	return nil
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

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}
	return true
}

func encodeHeader(name, content string) (string, error) {
	if isASCII(content) {
		if len(content) < 66 {
			return fmt.Sprintf("%s: %s\r\n", name, content), nil
		}
	}

	var ret string

	chunks := chunkString(content, 30)
	if len(chunks) == 1 {
		ret = chunks[0]
	} else {
		for i, chunk := range chunks {
			if i == 0 {
				ret += fmt.Sprintf("%s\r\n", encodeWord(chunk))
			} else {
				ret += fmt.Sprintf(" %s\r\n", encodeWord(chunk))
			}
		}
	}

	ret = fmt.Sprintf("%s: %s", name, ret)

	return ret, nil
}

func encodeWord(str string) string {
	return fmt.Sprintf("=?utf-8?B?%s?=", base64.StdEncoding.EncodeToString([]byte(str)))
}

func chunkString(s string, chunkSize int) []string {
	length := len(s)
	if chunkSize >= length {
		return []string{s}
	}
	var chunks []string = make([]string, 0, (length-1)/chunkSize+1)
	currentLen := 0
	currentStart := 0
	bits := strings.Split(s, "")
	for i := range bits {
		if currentLen == chunkSize {
			chunks = append(chunks, s[currentStart:i])
			currentLen = 0
			currentStart = i
		}
		currentLen++
	}
	chunks = append(chunks, s[currentStart:])
	return chunks
}
