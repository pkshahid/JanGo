package auth

// User represents an authenticated user in the system.
// This is a minimal interface to be expanded later.
type User interface {
	IsAuthenticated() bool
	GetUsername() string
}
