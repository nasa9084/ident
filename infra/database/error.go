package database

// Error is package-specific error type.
type Error string

// error constants
const (
	ErrUserNotFound Error = "user not found"
)

// Error implements error interface.
func (e Error) Error() string { return string(e) }
