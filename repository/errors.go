package repository

// Error is package specific error type.
type Error string

func (e Error) Error() string {
	return string(e)
}

// Error constants for this package.
const (
	ErrUserExists Error = "given user ID has been used"
)
