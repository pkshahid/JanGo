package email

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"
)

// EmailTuple represents an email to be sent in SendMassMail.
type EmailTuple struct {
	Subject string
	Message string
	From    string
	To      []string
}

// SendMail sends a simple email.
func SendMail(subject, message, from string, to []string) error {
	msg := NewEmailMessage(subject, message, from, to)
	return msg.Send()
}

// SendMassMail sends multiple emails via a single connection.
func SendMassMail(messages []EmailTuple) int {
	conn := GetDefaultBackend()
	if conn == nil {
		return 0
	}

	err := conn.Open()
	if err != nil {
		return 0
	}
	defer conn.Close()

	var emailMessages []*EmailMessage
	for _, m := range messages {
		msg := NewEmailMessage(m.Subject, m.Message, m.From, m.To)
		emailMessages = append(emailMessages, msg)
	}

	sent, _ := conn.SendMessages(emailMessages)
	return sent
}

// MailAdmins sends a message to the site administrators.
// For now, hardcoded "admin@example.com". In a real setup, it loads ADMINS from settings.
func MailAdmins(subject, message string) error {
	// Usually prefix with settings.EMAIL_SUBJECT_PREFIX
	fullSubject := "[GoDjango] " + subject
	msg := NewEmailMessage(fullSubject, message, "server@example.com", []string{"admin@example.com"})
	return msg.Send()
}

// MailManagers sends a message to the site managers.
func MailManagers(subject, message string) error {
	fullSubject := "[GoDjango] " + subject
	msg := NewEmailMessage(fullSubject, message, "server@example.com", []string{"manager@example.com"})
	return msg.Send()
}

// SendTemplatedMail renders subject and body from templates and sends the email.
// It looks for <templatePrefix>_subject.txt, <templatePrefix>_body.txt,
// and optionally <templatePrefix>_body.html
func SendTemplatedMail(templatePrefix string, context any, to []string, from string) error {
	subjectPath := templatePrefix + "_subject.txt"
	bodyTxtPath := templatePrefix + "_body.txt"
	bodyHtmlPath := templatePrefix + "_body.html"

	subjectRaw, err := os.ReadFile(subjectPath)
	if err != nil {
		return fmt.Errorf("could not read subject template: %v", err)
	}
	subjectStr := renderTemplateString(string(subjectRaw), context)
	// Remove newlines from subject
	subjectStr = strings.ReplaceAll(subjectStr, "\n", "")
	subjectStr = strings.ReplaceAll(subjectStr, "\r", "")

	bodyTxtRaw, err := os.ReadFile(bodyTxtPath)
	if err != nil {
		return fmt.Errorf("could not read body template: %v", err)
	}
	bodyTxtStr := renderTemplateString(string(bodyTxtRaw), context)

	htmlRaw, htmlErr := os.ReadFile(bodyHtmlPath)

	if htmlErr == nil {
		// We have an HTML alternative
		msg := NewEmailMultiAlternatives(subjectStr, bodyTxtStr, from, to)
		htmlStr := renderTemplateString(string(htmlRaw), context)
		msg.AttachAlternative(htmlStr, "text/html")
		return msg.Send()
	}

	// Just text
	msg := NewEmailMessage(subjectStr, bodyTxtStr, from, to)
	return msg.Send()
}

// renderTemplateString is a simple text/template renderer.
func renderTemplateString(tmplStr string, data any) string {
	tmpl, err := template.New("email").Parse(tmplStr)
	if err != nil {
		return tmplStr // Fallback if parse fails
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return tmplStr
	}
	return buf.String()
}
