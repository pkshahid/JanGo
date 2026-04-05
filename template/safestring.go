package template

// SafeString represents a string that has already been safely HTML-escaped
// and should bypass autoescaping during rendering.
type SafeString string
