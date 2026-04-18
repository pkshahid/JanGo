package email

import (
	"os"
	"strings"
	"testing"
)

func TestEmailMessage_Basic(t *testing.T) {
	OutBox = nil // Clear outbox
	SetDefaultBackend(&LocmemEmailBackend{})

	msg := NewEmailMessage("Test Subject", "Test Body", "from@example.com", []string{"to@example.com"})
	err := msg.Send()
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	if len(OutBox) != 1 {
		t.Fatalf("Expected 1 message in outbox, got %d", len(OutBox))
	}

	sentMsg := OutBox[0]
	if sentMsg.Subject != "Test Subject" {
		t.Errorf("Expected subject 'Test Subject', got %s", sentMsg.Subject)
	}
	if sentMsg.From != "from@example.com" {
		t.Errorf("Expected from 'from@example.com', got %s", sentMsg.From)
	}
	if len(sentMsg.To) != 1 || sentMsg.To[0] != "to@example.com" {
		t.Errorf("Expected to 'to@example.com'")
	}
}

func TestEmailMessage_Bytes(t *testing.T) {
	msg := NewEmailMessage("Subj", "Body", "from@a.com", []string{"to@a.com"})
	b, err := msg.Message()
	if err != nil {
		t.Fatalf("Failed to construct message bytes: %v", err)
	}

	str := string(b)
	if !strings.Contains(str, "Subject: =?utf-8?q?Subj?=") && !strings.Contains(str, "Subject: Subj") {
		t.Errorf("Subject not encoded correctly in bytes. Got: %s", str)
	}
	if !strings.Contains(str, "From: from@a.com") {
		t.Errorf("From missing from bytes")
	}
	if !strings.Contains(str, "To: to@a.com") {
		t.Errorf("To missing from bytes")
	}
	if !strings.Contains(str, "Body") {
		t.Errorf("Body missing from bytes")
	}
}

func TestEmailMultiAlternatives(t *testing.T) {
	OutBox = nil
	SetDefaultBackend(&LocmemEmailBackend{})

	msg := NewEmailMultiAlternatives("Multi", "Text Body", "f@a.com", []string{"t@a.com"})
	msg.AttachAlternative("<h1>HTML Body</h1>", "text/html")
	err := msg.Send()
	if err != nil {
		t.Fatalf("Failed to send: %v", err)
	}

	if len(OutBox) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(OutBox))
	}
	sent := OutBox[0]
	if len(sent.Alternatives) != 1 {
		t.Errorf("Expected 1 alternative, got %d", len(sent.Alternatives))
	}

	b, _ := sent.Message()
	str := string(b)
	if !strings.Contains(str, "Text Body") {
		t.Errorf("Text body missing")
	}
	if !strings.Contains(str, "<h1>HTML Body</h1>") {
		t.Errorf("HTML body missing")
	}
}

func TestShortcuts(t *testing.T) {
	OutBox = nil
	SetDefaultBackend(&LocmemEmailBackend{})

	err := SendMail("S1", "M1", "f@a.com", []string{"t@a.com"})
	if err != nil {
		t.Errorf("SendMail failed")
	}

	sent := SendMassMail([]EmailTuple{
		{"S2", "M2", "f@a.com", []string{"t1@a.com"}},
		{"S3", "M3", "f@a.com", []string{"t2@a.com"}},
	})
	if sent != 2 {
		t.Errorf("SendMassMail expected 2, got %d", sent)
	}

	if len(OutBox) != 3 {
		t.Errorf("Expected 3 in outbox, got %d", len(OutBox))
	}
}

func TestSendTemplatedMail(t *testing.T) {
	OutBox = nil
	SetDefaultBackend(&LocmemEmailBackend{})

	// Create dummy templates
	os.WriteFile("test_email_subject.txt", []byte("Hello {{.Name}}\n"), 0644)
	os.WriteFile("test_email_body.txt", []byte("Welcome {{.Name}} to our site."), 0644)

	defer os.Remove("test_email_subject.txt")
	defer os.Remove("test_email_body.txt")

	ctx := struct{ Name string }{"Bob"}
	err := SendTemplatedMail("test_email", ctx, []string{"b@b.com"}, "from@b.com")
	if err != nil {
		t.Fatalf("SendTemplatedMail failed: %v", err)
	}

	if len(OutBox) != 1 {
		t.Fatalf("Expected 1 message in outbox")
	}
	msg := OutBox[0]
	if msg.Subject != "Hello Bob" { // Note: newline should be stripped
		t.Errorf("Expected 'Hello Bob', got %q", msg.Subject)
	}
	if msg.Body != "Welcome Bob to our site." {
		t.Errorf("Expected 'Welcome Bob...', got %q", msg.Body)
	}
}

func TestAsyncEmail(t *testing.T) {
	OutBox = nil
	SetDefaultBackend(&LocmemEmailBackend{})

	msg := NewEmailMessage("Subj", "Body", "f@a.com", []string{"t@a.com"})
	errChan := msg.SendAsync()

	err := <-errChan
	if err != nil {
		t.Errorf("Async send failed")
	}
	if len(OutBox) != 1 {
		t.Errorf("Expected 1 message in outbox")
	}
}
