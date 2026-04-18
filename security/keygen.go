package security

import (
	"crypto/rand"
	"math/big"
)

// GenerateSecretKey generates a cryptographically secure 50-character random string.
// It uses the same character set as Django: 'abcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*(-_=+)'
func GenerateSecretKey() (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*(-_=+)"
	length := 50
	b := make([]byte, length)
	max := big.NewInt(int64(len(charset)))

	for i := range b {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		b[i] = charset[n.Int64()]
	}
	return string(b), nil
}
