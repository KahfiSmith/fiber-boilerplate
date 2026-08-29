package mail

import (
	"context"
	"fmt"
	"log/slog"
)

type Mailer interface {
	SendPasswordReset(ctx context.Context, toEmail string, resetToken string) error
	SendEmailVerification(ctx context.Context, toEmail string, verificationToken string) error
}

type LogMailer struct {
	appName string
}

func NewLogMailer(appName string) *LogMailer {
	return &LogMailer{appName: appName}
}

func (m *LogMailer) SendPasswordReset(ctx context.Context, toEmail string, resetToken string) error {
	slog.Info("email_delivery_simulated",
		slog.String("type", "password_reset"),
		slog.String("to", toEmail),
		slog.String("token", resetToken),
	)
	return nil
}

func (m *LogMailer) SendEmailVerification(ctx context.Context, toEmail string, verificationToken string) error {
	slog.Info("email_delivery_simulated",
		slog.String("type", "email_verification"),
		slog.String("to", toEmail),
		slog.String("token", verificationToken),
	)
	return nil
}

type SMTPMailer struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

func NewSMTPMailer(host string, port int, user string, pass string, from string) *SMTPMailer {
	return &SMTPMailer{
		Host:     host,
		Port:     port,
		Username: user,
		Password: pass,
		From:     from,
	}
}

func (m *SMTPMailer) SendPasswordReset(ctx context.Context, toEmail string, resetToken string) error {
	return fmt.Errorf("smtp delivery not configured")
}

func (m *SMTPMailer) SendEmailVerification(ctx context.Context, toEmail string, verificationToken string) error {
	return fmt.Errorf("smtp delivery not configured")
}
