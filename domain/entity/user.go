package entity

// User entity object.
type User struct {
	ID         string
	Password   string
	TOTPSecret string
	Email      string

	TOTPVerified bool
}
