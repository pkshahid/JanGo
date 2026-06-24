package email

import (
	"bytes"
	"fmt"
	"mime"
	"mime/multipart"
	"net/textproto"
	"strings"
)

// Attachment represents an email attachment.
type Attachment struct {
	Filename string
	Content  []byte
	Mimetype string
}

// Alternative represents an alternative body format (like HTML).
type Alternative struct {
	Content  string
	Mimetype string
}

// EmailMessage represents a standard email message.
type EmailMessage struct {
	Subject     string
	Body        string
	From        string
	To          []string
	Bcc         []string
	Cc          []string
	ReplyTo     []string
	Headers     map[string]string
	Attachments []Attachment

	Alternatives []Alternative // Used internally by EmailMultiAlternatives

	connection EmailBackend
}

// NewEmailMessage creates a new EmailMessage.
func NewEmailMessage(subject, body, from string, to []string) *EmailMessage {
	return &EmailMessage{
		Subject: subject,
		Body:    body,
		From:    from,
		To:      to,
		Headers: make(map[string]string),
	}
}

// SetConnection explicitly sets the backend for sending.
func (e *EmailMessage) SetConnection(backend EmailBackend) {
	e.connection = backend
}

// GetConnection returns the current connection or the default.
func (e *EmailMessage) GetConnection() EmailBackend {
	if e.connection != nil {
		return e.connection
	}
	return GetDefaultBackend()
}

// Send sends the email using the configured backend.
func (e *EmailMessage) Send() error {
	conn := e.GetConnection()
	if conn == nil {
		return fmt.Errorf("no email backend configured")
	}

	// Open connection if not already open (Backends usually handle this or we explicitly open)
	err := conn.Open()
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.SendMessages([]*EmailMessage{e})
	return err
}

// SendAsync sends the email via a goroutine and returns a channel for the error.
func (e *EmailMessage) SendAsync() <-chan error {
	ch := make(chan error, 1)
	go func() {
		ch <- e.Send()
		close(ch)
	}()
	return ch
}

// Message constructs the raw RFC 822 email bytes.
func (e *EmailMessage) Message() ([]byte, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// In a full implementation, we'd correctly format Date, Message-ID, etc.
	// We'll write the basic headers here.
	headers := make(map[string]string)

	// Copy explicit headers
	for k, v := range e.Headers {
		headers[k] = v
	}

	// Set standard headers
	headers["From"] = e.From
	if len(e.To) > 0 {
		headers["To"] = strings.Join(e.To, ", ")
	}
	if len(e.Cc) > 0 {
		headers["Cc"] = strings.Join(e.Cc, ", ")
	}
	if len(e.ReplyTo) > 0 {
		headers["Reply-To"] = strings.Join(e.ReplyTo, ", ")
	}
	// Note: Bcc is usually handled by the SMTP envelope, not the headers,
	// but some implementations strip it. We'll omit Bcc from headers.

	// Encode Subject
	encodedSubject := mime.QEncoding.Encode("utf-8", e.Subject)
	headers["Subject"] = encodedSubject
	headers["MIME-Version"] = "1.0"

	hasAttachments := len(e.Attachments) > 0

	// To cleanly handle alternatives, we'll check a hidden or exported field.
	var alternatives []Alternative
	if e.Alternatives != nil {
		alternatives = e.Alternatives
	}

	hasAlternatives := len(alternatives) > 0

	var rootWriter *multipart.Writer
	var bodyWriter *multipart.Writer

	if hasAttachments {
		rootWriter = writer
		headers["Content-Type"] = fmt.Sprintf("multipart/mixed; boundary=%q", rootWriter.Boundary())
	} else if hasAlternatives {
		rootWriter = writer
		headers["Content-Type"] = fmt.Sprintf("multipart/alternative; boundary=%q", rootWriter.Boundary())
	} else {
		// Just text
		headers["Content-Type"] = "text/plain; charset=\"utf-8\""
	}

	// Write headers
	for k, v := range headers {
		fmt.Fprintf(&buf, "%s: %s\r\n", k, v)
	}
	fmt.Fprintf(&buf, "\r\n") // End of headers

	// Write body
	if rootWriter == nil {
		buf.WriteString(e.Body)
	} else {
		if hasAttachments && hasAlternatives {
			// Mixed -> Alternative -> text, html
			bodyBuf := &bytes.Buffer{}
			bodyWriter = multipart.NewWriter(bodyBuf)

			altPartHeader := make(textproto.MIMEHeader)
			altPartHeader.Set("Content-Type", fmt.Sprintf("multipart/alternative; boundary=%q", bodyWriter.Boundary()))

			part, err := rootWriter.CreatePart(altPartHeader)
			if err != nil {
				return nil, err
			}

			// We will write into `part` after we build the alternative body

			// Build Alternative part
			writeTextPart(bodyWriter, e.Body)
			for _, alt := range alternatives {
				writeAlternativePart(bodyWriter, alt)
			}
			bodyWriter.Close()
			part.Write(bodyBuf.Bytes())

		} else if hasAlternatives {
			writeTextPart(rootWriter, e.Body)
			for _, alt := range alternatives {
				writeAlternativePart(rootWriter, alt)
			}
		} else if hasAttachments {
			writeTextPart(rootWriter, e.Body)
		}

		// Write Attachments
		for _, att := range e.Attachments {
			attHeader := make(textproto.MIMEHeader)
			attHeader.Set("Content-Type", att.Mimetype)
			attHeader.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, att.Filename))
			attHeader.Set("Content-Transfer-Encoding", "base64")

			part, err := rootWriter.CreatePart(attHeader)
			if err != nil {
				return nil, err
			}
			// In a real impl, we'd base64 encode here, for simplicity we just write
			part.Write(att.Content)
		}

		rootWriter.Close()
	}

	return buf.Bytes(), nil
}

func writeTextPart(w *multipart.Writer, body string) {
	h := make(textproto.MIMEHeader)
	h.Set("Content-Type", "text/plain; charset=\"utf-8\"")
	part, _ := w.CreatePart(h)
	part.Write([]byte(body))
}

func writeAlternativePart(w *multipart.Writer, alt Alternative) {
	h := make(textproto.MIMEHeader)
	h.Set("Content-Type", fmt.Sprintf("%s; charset=\"utf-8\"", alt.Mimetype))
	part, _ := w.CreatePart(h)
	part.Write([]byte(alt.Content))
}

// Recipients returns all recipients (To, Cc, Bcc)
func (e *EmailMessage) Recipients() []string {
	var rec []string
	rec = append(rec, e.To...)
	rec = append(rec, e.Cc...)
	rec = append(rec, e.Bcc...)
	return rec
}

// EmailMultiAlternatives extends EmailMessage to support alternative formats like HTML.
type EmailMultiAlternatives struct {
	*EmailMessage
}

// NewEmailMultiAlternatives creates a new EmailMultiAlternatives.
func NewEmailMultiAlternatives(subject, body, from string, to []string) *EmailMultiAlternatives {
	return &EmailMultiAlternatives{
		EmailMessage: &EmailMessage{
			Subject: subject,
			Body:    body,
			From:    from,
			To:      to,
			Headers: make(map[string]string),
		},
	}
}

// AttachAlternative adds an alternative content part.
func (e *EmailMultiAlternatives) AttachAlternative(content, mimetype string) {
	e.EmailMessage.Alternatives = append(e.EmailMessage.Alternatives, Alternative{
		Content:  content,
		Mimetype: mimetype,
	})
}
