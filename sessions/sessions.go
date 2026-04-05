package sessions

// Session represents a user session.
// This is a minimal interface to be expanded later.
type Session interface {
	Get(key string) any
	Set(key string, value any)
	Delete(key string)
}
