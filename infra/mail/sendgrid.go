package mail

import (
	"github.com/nasa9084/ident/domain/service"
	sendgrid "github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// sgmail is an implementation of service.Mail interace.
// this struct uses SendGrid email delivery service as background.
type sgmail struct {
	client *sendgrid.Client
	from   *mail.Email
}

// NewSendGrid returns a new sendgrid client as service.Mail.
func NewSendGrid(apikey, from string) service.Mail {
	return &sgmail{
		client: sendgrid.NewSendClient(apikey),
		from:   &mail.Email{Address: from},
	}
}

func (sg *sgmail) Send(to, subject, body string) error {
	msg := mail.NewV3MailInit(
		/* from */ sg.from,
		/* subject */ subject,
		/* to */ &mail.Email{Address: to},
		/* content */ mail.NewContent("text/plain", body),
	)
	_, err := sg.client.Send(msg)
	return err
}
