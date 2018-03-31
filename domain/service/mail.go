package service

// Mail interface represents e-mail client.
type Mail interface {
	Send(to, subject, body string) error
}
