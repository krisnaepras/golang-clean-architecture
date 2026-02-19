package email

import (
	"fmt"

	gomail "github.com/wneessen/go-mail"
)

// Mailer wraps go-mail to send transactional emails.
type Mailer struct {
	host     string
	port     int
	username string
	password string
	from     string
	fromName string
	appName  string
}

// NewMailer creates a Mailer from explicit parameters.
func NewMailer(host string, port int, username, password, from, fromName, appName string) *Mailer {
	return &Mailer{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
		fromName: fromName,
		appName:  appName,
	}
}

// SendOTP sends an OTP email for the given purpose to the target address.
// Returns an error if sending fails; the caller decides whether to propagate.
func (m *Mailer) SendOTP(to, otpCode, purpose string) error {
	htmlBody, err := BuildOTPEmailHTML(m.appName, otpCode, purpose)
	if err != nil {
		return fmt.Errorf("build OTP HTML: %w", err)
	}

	subject := m.subjectForPurpose(purpose)

	msg := gomail.NewMsg()
	if err := msg.FromFormat(m.fromName, m.from); err != nil {
		return fmt.Errorf("set from address: %w", err)
	}
	if err := msg.To(to); err != nil {
		return fmt.Errorf("set to address: %w", err)
	}
	msg.Subject(subject)
	msg.SetBodyString(gomail.TypeTextHTML, htmlBody)

	opts := []gomail.Option{
		gomail.WithPort(m.port),
		gomail.WithTLSPolicy(gomail.TLSOpportunistic),
	}
	if m.username != "" {
		opts = append(opts,
			gomail.WithSMTPAuth(gomail.SMTPAuthPlain),
			gomail.WithUsername(m.username),
			gomail.WithPassword(m.password),
		)
	}

	client, err := gomail.NewClient(m.host, opts...)
	if err != nil {
		return fmt.Errorf("create mail client: %w", err)
	}

	if err := client.DialAndSend(msg); err != nil {
		return fmt.Errorf("send email: %w", err)
	}
	return nil
}

func (m *Mailer) subjectForPurpose(purpose string) string {
	switch purpose {
	case "email_verification":
		return "[" + m.appName + "] Kode Verifikasi Email"
	case "reset_password":
		return "[" + m.appName + "] Kode Reset Password"
	case "login":
		return "[" + m.appName + "] Kode Login OTP"
	default:
		return "[" + m.appName + "] Kode OTP"
	}
}
