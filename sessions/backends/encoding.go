package backends

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/godjango/godjango/core/settings"
)

// Sign encrypts and signs session data into a base64 string.
func Sign(data map[string]any) (string, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	b64Data := base64.RawURLEncoding.EncodeToString(jsonData)
	signature := makeSignature(b64Data)

	return fmt.Sprintf("%s.%s", b64Data, signature), nil
}

// Unsign verifies the signature and decodes the session data.
func Unsign(signedData string) (map[string]any, error) {
	parts := strings.Split(signedData, ".")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid signed session format")
	}

	b64Data, signature := parts[0], parts[1]

	expectedSig := makeSignature(b64Data)
	if signature != expectedSig {
		return nil, fmt.Errorf("signature verification failed")
	}

	jsonData, err := base64.RawURLEncoding.DecodeString(b64Data)
	if err != nil {
		return nil, err
	}

	var data map[string]any
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, err
	}

	return data, nil
}

func makeSignature(data string) string {
	s := settings.Get()
	mac := hmac.New(sha256.New, []byte(s.SECRET_KEY))
	mac.Write([]byte(data))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
