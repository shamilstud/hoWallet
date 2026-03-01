package service

import (
	"fmt"
	"net/smtp"
	"strings"

	"github.com/howallet/howallet/internal/config"
)

// EmailService sends transactional emails via SMTP.
type EmailService struct {
	cfg *config.SMTPConfig
}

func NewEmailService(cfg *config.SMTPConfig) *EmailService {
	return &EmailService{cfg: cfg}
}

// SendInvitation sends a household invitation email with a link to accept.
func (s *EmailService) SendInvitation(toEmail, householdName, inviterName, token, frontendURL string) error {
	acceptURL := fmt.Sprintf("%s/invite/%s", strings.TrimRight(frontendURL, "/"), token)

	subject := fmt.Sprintf("You've been invited to join \"%s\" on hoWallet", householdName)
	body := fmt.Sprintf(`Hello!

%s has invited you to join the household "%s" on hoWallet.

Click the link below to accept the invitation:
%s

This invitation will expire in 7 days.

If you don't have a hoWallet account yet, please register first and then use the link above.

— hoWallet Team
`, inviterName, householdName, acceptURL)

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		s.cfg.From, toEmail, subject, body)

	auth := smtp.PlainAuth("", s.cfg.User, s.cfg.Password, s.cfg.Host)

	addr := fmt.Sprintf("%s:%s", s.cfg.Host, s.cfg.Port)
	return smtp.SendMail(addr, auth, s.cfg.From, []string{toEmail}, []byte(msg))
}
