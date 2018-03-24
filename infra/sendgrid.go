package infra

import "github.com/sendgrid/sendgrid-go/helpers/mail"

var frommail = &mail.Email{
	Address: "foo@bar",
}

func NewSGMail(to, sub, body string) *mail.SGMailV3 {
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

func NewVerificationMail(to, sessid string) *mail.SGMailV3 {
	body := `http://localhost:8080/v1/user/email/` + sessid
	return NewSGMail(to, "verify your mail address", body)
}
