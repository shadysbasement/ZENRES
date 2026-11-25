package auth

import (
	"context"
	"fmt"
	"log/slog"
	"net/smtp"
	"strings"

	"github.com/ProtonMail/gopenpgp/v3/crypto"
)

// Mailer delivers transactional email. If recipientKey is set, the body is PGP-encrypted
// end-to-end (Proton WKD flow) before it leaves the server.
type Mailer interface {
	Send(ctx context.Context, to, subject, body string, recipientKey *crypto.Key) error
}

// LogMailer prints links to the server log — for local development only.
type LogMailer struct{}

func (LogMailer) Send(_ context.Context, to, subject, body string, k *crypto.Key) error {
	slog.Info("mail (dev log)", "to", to, "subject", subject, "pgp_encrypted", k != nil, "body", body)
	return nil
}

// SMTPMailer sends via an authenticated SMTP relay (Postmark/SES/etc.).
type SMTPMailer struct {
	Addr, From, User, Pass string
}

func (m SMTPMailer) Send(_ context.Context, to, subject, body string, k *crypto.Key) error {
	if k != nil {
		enc, err := crypto.PGP().Encryption().Recipient(k).New()
		if err != nil {
			return err
		}
		msg, err := enc.Encrypt([]byte(body))
		if err != nil {
			return err
		}
		if body, err = msg.Armor(); err != nil {
			return err
		}
	}
	host := strings.Split(m.Addr, ":")[0]
	raw := fmt.Sprintf("From: Zevra <%s>\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n%s", m.From, to, subject, body)
	return smtp.SendMail(m.Addr, smtp.PlainAuth("", m.User, m.Pass, host), m.From, []string{to}, []byte(raw))
}
