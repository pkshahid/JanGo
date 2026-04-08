package email

// Global Outbox for Locmem backend
var OutBox []*EmailMessage

// Default backend logic
var defaultBackend EmailBackend

// GetDefaultBackend gets or initializes the default email backend.
func GetDefaultBackend() EmailBackend {
	if defaultBackend == nil {
		// Just for default test purposes, or we could load from settings.
		// Usually Locmem or SMTP depending on settings.
		defaultBackend = &LocmemEmailBackend{}
	}
	return defaultBackend
}

// SetDefaultBackend explicitly sets the backend.
func SetDefaultBackend(backend EmailBackend) {
	defaultBackend = backend
}
