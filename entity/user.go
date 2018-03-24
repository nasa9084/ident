package entity

type User struct {
	ID         string
	Password   string
	TOTPSecret string
	Email      string

	TOTPVerified bool
}
