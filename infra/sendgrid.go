package infra

import "github.com/sendgrid/sendgrid-go/helpers/mail"

// NewSGMail creates mail object from TO, SUBJECT, and message BODY.
func NewSGMail(from, to, sub, body string) *mail.SGMailV3 {
	frommail := &mail.Email{
		Address: from,
	}
	tomail := &mail.Email{
		Address: to,
	}
	m := mail.NewSingleEmail(
		/* from */ frommail,
		/* subject */ sub,
		/* to */ tomail,
		/* plainbody */ body,
		/* htmlbody */ body,
	)
	return m
}

// NewVerificationMail creates a new email-verification mail object.
func NewVerificationMail(from, to, sessid string) *mail.SGMailV3 {
	body := `http://localhost:8080/v1/user/email/` + sessid
	return NewSGMail(from, to, "verify your mail address", body)
}
