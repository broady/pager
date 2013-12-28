package pager

import (
	"fmt"

	"appengine"
	"appengine/mail"
)

const emailTemplate = `
==================================
Respond with "ACK" to acknowledge.
==================================

%s
`

func sendMail(c appengine.Context, email string, esc *Escalation) error {
	c.Debugf("mail to %s", email)
	msg := &mail.Message{
		Sender:  fmt.Sprintf("Pager <acknowledge+%s@%s.appspotmail.com>", esc.key.Encode(), appengine.AppID(c)),
		To:      []string{email},
		Subject: fmt.Sprintf("PAGE: %s", esc.Subject),
		Body:    fmt.Sprintf(emailTemplate, esc.Body),
	}
	return mail.Send(c, msg)
}

func sendPhone(c appengine.Context, phone string, esc *Escalation) error {
	c.Debugf("calling %s", phone)
	return nil
}

func sendSMS(c appengine.Context, phone string, esc *Escalation) error {
	c.Debugf("SMS to %s", phone)
	return nil
}
