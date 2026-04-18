package exceptions

// Http404 represents a resource not found error.
type Http404 struct {
	Message string
}

func (e *Http404) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return "Not Found"
}

// PermissionDenied represents an unauthorized access error.
type PermissionDenied struct {
	Message string
}

func (e *PermissionDenied) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return "Permission Denied"
}

// SuspiciousOperation represents a suspicious user operation (e.g. CSRF failure).
type SuspiciousOperation struct {
	Message string
}

func (e *SuspiciousOperation) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return "Suspicious Operation"
}

// BadRequest represents a 400 Bad Request error.
type BadRequest struct {
	Message string
}

func (e *BadRequest) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return "Bad Request"
}

// DisallowedHost represents an invalid HTTP Host header.
type DisallowedHost struct {
	Message string
}

func (e *DisallowedHost) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return "Disallowed Host"
}

// RequestDataTooBig represents an uploaded file or request body that exceeds limits.
type RequestDataTooBig struct {
	Message string
}

func (e *RequestDataTooBig) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return "Request Data Too Big"
}
