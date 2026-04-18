package email

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// EmailBackend is the interface for all email sending implementations.
type EmailBackend interface {
	Open() error
	Close() error
	SendMessages(messages []*EmailMessage) (int, error)
}

// SMTPEmailBackend sends emails via SMTP.
type SMTPEmailBackend struct {
	Host     string
	Port     int
	Username string
	Password string
	UseTLS   bool
	UseSSL   bool

	client *smtp.Client
	mu     sync.Mutex
}

func (b *SMTPEmailBackend) Open() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.client != nil {
		return nil
	}

	addr := fmt.Sprintf("%s:%d", b.Host, b.Port)

	var c *smtp.Client
	var err error

	if b.UseSSL {
		tlsconfig := &tls.Config{
			ServerName: b.Host,
		}
		conn, err := tls.Dial("tcp", addr, tlsconfig)
		if err != nil {
			return err
		}
		c, err = smtp.NewClient(conn, b.Host)
		if err != nil {
			return err
		}
	} else {
		c, err = smtp.Dial(addr)
		if err != nil {
			return err
		}

		if b.UseTLS {
			if err = c.StartTLS(&tls.Config{ServerName: b.Host}); err != nil {
				return err
			}
		}
	}

	if b.Username != "" && b.Password != "" {
		auth := smtp.PlainAuth("", b.Username, b.Password, b.Host)
		if err = c.Auth(auth); err != nil {
			return err
		}
	}

	b.client = c
	return nil
}

func (b *SMTPEmailBackend) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.client != nil {
		err := b.client.Quit()
		b.client = nil
		return err
	}
	return nil
}

func (b *SMTPEmailBackend) SendMessages(messages []*EmailMessage) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.client == nil {
		return 0, fmt.Errorf("SMTP connection not open")
	}

	sent := 0
	for _, msg := range messages {
		if err := b.client.Mail(msg.From); err != nil {
			return sent, err
		}

		for _, rcpt := range msg.Recipients() {
			if err := b.client.Rcpt(rcpt); err != nil {
				return sent, err
			}
		}

		w, err := b.client.Data()
		if err != nil {
			return sent, err
		}

		rawMsg, err := msg.Message()
		if err != nil {
			return sent, err
		}

		_, err = w.Write(rawMsg)
		if err != nil {
			return sent, err
		}

		err = w.Close()
		if err != nil {
			return sent, err
		}
		sent++
	}

	return sent, nil
}


// ConsoleEmailBackend writes emails to standard output.
type ConsoleEmailBackend struct {
	mu sync.Mutex
}

func (b *ConsoleEmailBackend) Open() error { return nil }
func (b *ConsoleEmailBackend) Close() error { return nil }

func (b *ConsoleEmailBackend) SendMessages(messages []*EmailMessage) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	sent := 0
	for _, msg := range messages {
		raw, err := msg.Message()
		if err == nil {
			fmt.Println(strings.Repeat("-", 79))
			fmt.Println(string(raw))
			fmt.Println(strings.Repeat("-", 79))
			sent++
		}
	}
	return sent, nil
}


// FileEmailBackend writes emails to files.
type FileEmailBackend struct {
	FilePath string
	mu       sync.Mutex
}

func (b *FileEmailBackend) Open() error {
	if b.FilePath == "" {
		return fmt.Errorf("FileEmailBackend requires a FilePath")
	}
	return os.MkdirAll(b.FilePath, 0755)
}

func (b *FileEmailBackend) Close() error { return nil }

func (b *FileEmailBackend) SendMessages(messages []*EmailMessage) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	sent := 0
	for _, msg := range messages {
		raw, err := msg.Message()
		if err != nil {
			continue
		}
		filename := filepath.Join(b.FilePath, fmt.Sprintf("%d-%d.log", time.Now().UnixNano(), sent))
		if err := os.WriteFile(filename, raw, 0644); err == nil {
			sent++
		}
	}
	return sent, nil
}


// LocmemEmailBackend appends emails to the global email.OutBox slice for testing.
type LocmemEmailBackend struct {
	mu sync.Mutex
}

func (b *LocmemEmailBackend) Open() error { return nil }
func (b *LocmemEmailBackend) Close() error { return nil }

func (b *LocmemEmailBackend) SendMessages(messages []*EmailMessage) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, msg := range messages {
		OutBox = append(OutBox, msg)
	}
	return len(messages), nil
}


// DummyEmailBackend does absolutely nothing.
type DummyEmailBackend struct{}

func (b *DummyEmailBackend) Open() error { return nil }
func (b *DummyEmailBackend) Close() error { return nil }

func (b *DummyEmailBackend) SendMessages(messages []*EmailMessage) (int, error) {
	return len(messages), nil
}
