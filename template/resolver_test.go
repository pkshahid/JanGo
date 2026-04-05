package template

import (
	"testing"
)

type TestProfile struct {
	Name string
}

func (p TestProfile) GetUpperName() string {
	return "UPPER_" + p.Name
}

type TestUser struct {
	Profile *TestProfile
	IsAdmin bool
}

func TestResolveVariable(t *testing.T) {
	ctx := NewContext(map[string]any{
		"dict": map[string]any{
			"key": "value",
		},
		"user": &TestUser{
			Profile: &TestProfile{Name: "Alice"},
			IsAdmin: true,
		},
	})

	tests := []struct {
		Path     string
		Expected any
	}{
		{`"literal"`, "literal"},
		{`'literal2'`, "literal2"},
		{"dict.key", "value"},
		{"user.IsAdmin", true},
		{"user.isAdmin", true}, // lowercase fallback
		{"user.Profile.Name", "Alice"},
		{"user.profile.name", "Alice"},
		{"user.profile.GetUpperName", "UPPER_Alice"},
		{"user.profile.getUpperName", "UPPER_Alice"},
		{"missing", ""},
		{"user.missing", ""},
	}

	for _, test := range tests {
		val := resolveVariable(test.Path, ctx)
		if val != test.Expected {
			t.Errorf("Path %q: Expected %v, got %v", test.Path, test.Expected, val)
		}
	}
}
