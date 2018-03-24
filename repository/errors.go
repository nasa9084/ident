package repository

type Error string

func (e Error) Error() string {
	return string(e)
}

const (
	ErrUserExists Error = "given user ID has been used"
)
