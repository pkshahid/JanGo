package hashers

import (
	"testing"
)

func TestArgon2idHasher(t *testing.T) {
	password := "my_secret_password"

	encoded := MakePassword(password)

	if !CheckPassword(password, encoded) {
		t.Errorf("Argon2id verification failed")
	}
	if CheckPassword("wrong_password", encoded) {
		t.Errorf("Argon2id verified wrong password")
	}
}

func TestBCryptHasher(t *testing.T) {
	hasher := &BCryptHasher{}
	password := "my_secret_password"

	encoded := hasher.Encode(password, "")

	if !CheckPassword(password, encoded) {
		t.Errorf("BCrypt verification failed")
	}
	if CheckPassword("wrong_password", encoded) {
		t.Errorf("BCrypt verified wrong password")
	}

	// Test legacy non-prefixed verify
	bareHash := encoded[7:] // remove 'bcrypt$'
	if !CheckPassword(password, bareHash) {
		t.Errorf("BCrypt legacy verification failed")
	}
}

func TestIsPasswordUsable(t *testing.T) {
	if IsPasswordUsable("!unusable") {
		t.Errorf("Expected false for unusable password")
	}
	if !IsPasswordUsable("argon2id$v=19$...") {
		t.Errorf("Expected true for valid hash format")
	}
	if IsPasswordUsable("") {
		t.Errorf("Expected false for empty string")
	}
}
