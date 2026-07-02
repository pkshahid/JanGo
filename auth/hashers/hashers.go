package hashers

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
)

// Hasher interface defines password hashing and verification.
type Hasher interface {
	Encode(password string, salt string) string
	Verify(password string, encoded string) bool
	Algorithm() string
}

var (
	DefaultHasher = &Argon2idHasher{}
	hashers       = map[string]Hasher{
		"argon2id": DefaultHasher,
		"bcrypt":   &BCryptHasher{},
	}
)

// MakePassword hashes a plain text password using the default hasher.
func MakePassword(password string) string {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		panic(err)
	}
	saltStr := base64.RawStdEncoding.EncodeToString(salt)
	return DefaultHasher.Encode(password, saltStr)
}

// CheckPassword verifies a password against an encoded hash.
func CheckPassword(password, encoded string) bool {
	if encoded == "" {
		return false
	}
	parts := strings.Split(encoded, "$")
	if len(parts) == 0 {
		return false
	}
	algo := parts[0]

	// Legacy un-prefixed hash support (fallback to bcrypt if it looks like one)
	if strings.HasPrefix(encoded, "$2a$") || strings.HasPrefix(encoded, "$2b$") {
		algo = "bcrypt"
	}

	hasher, ok := hashers[algo]
	if !ok {
		return false
	}
	return hasher.Verify(password, encoded)
}

// IsPasswordUsable checks if the encoded password is set to a non-usable value (e.g. `!`).
func IsPasswordUsable(encoded string) bool {
	return encoded != "" && !strings.HasPrefix(encoded, "!")
}

// Argon2idHasher implementation
type Argon2idHasher struct {
	Time    uint32
	Memory  uint32
	Threads uint8
	KeyLen  uint32
}

func (h *Argon2idHasher) Algorithm() string { return "argon2id" }

func (h *Argon2idHasher) Encode(password string, salt string) string {
	time := h.Time
	if time == 0 {
		time = 1
	}
	memory := h.Memory
	if memory == 0 {
		memory = 64 * 1024
	} // 64MB
	threads := h.Threads
	if threads == 0 {
		threads = 4
	}
	keyLen := h.KeyLen
	if keyLen == 0 {
		keyLen = 32
	}

	hash := argon2.IDKey([]byte(password), []byte(salt), time, memory, threads, keyLen)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf("argon2id$v=19$m=%d,t=%d,p=%d$%s$%s", memory, time, threads, salt, b64Hash)
}

func (h *Argon2idHasher) Verify(password string, encoded string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 5 {
		return false
	}

	// parse params: m=65536,t=1,p=4
	var memory, time uint32
	var threads uint8
	if _, err := fmt.Sscanf(parts[2], "m=%d,t=%d,p=%d", &memory, &time, &threads); err != nil {
		return false
	}

	salt := parts[3]
	hashStr := parts[4]

	hashBytes, err := base64.RawStdEncoding.DecodeString(hashStr)
	if err != nil {
		return false
	}

	keyLen := uint32(len(hashBytes))

	compareHash := argon2.IDKey([]byte(password), []byte(salt), time, memory, threads, keyLen)
	return subtle.ConstantTimeCompare(hashBytes, compareHash) == 1
}

// BCryptHasher implementation
type BCryptHasher struct{}

func (h *BCryptHasher) Algorithm() string { return "bcrypt" }

func (h *BCryptHasher) Encode(password string, salt string) string {
	// bcrypt manages its own salt natively.
	// We ignore the passed salt but comply with the interface.
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err) // Shouldn't happen unless max length exceeded
	}
	return fmt.Sprintf("bcrypt$%s", string(hash))
}

func (h *BCryptHasher) Verify(password string, encoded string) bool {
	// The encoded string might be `bcrypt$...` or just standard bcrypt output
	hash := encoded
	if strings.HasPrefix(encoded, "bcrypt$") {
		hash = encoded[7:]
	}
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// PBKDF2Hasher (legacy support stub)
type PBKDF2Hasher struct{}

func (h *PBKDF2Hasher) Algorithm() string                    { return "pbkdf2" }
func (h *PBKDF2Hasher) Encode(password, salt string) string  { return "" }
func (h *PBKDF2Hasher) Verify(password, encoded string) bool { return false }

func init() {
	hashers["pbkdf2"] = &PBKDF2Hasher{}
}
