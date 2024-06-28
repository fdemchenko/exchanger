package mailer

import (
	"bytes"
	"context"
	"text/template"
	"time"

	"github.com/fdemchenko/exchanger/web/templates"
	"github.com/go-mail/mail/v2"
	"github.com/rs/zerolog/log"
)

const (
	DialerTimeout = 5 * time.Second
)

type mailer struct {
	dialer         *mail.Dialer
	sender         string
	emailService   EmailService
	rateService    RateService
	updateInterval time.Duration
	stopChan       chan bool
}

type RateService interface {
	GetRate(context.Context, string) (float32, error)
}

type EmailService interface {
	Create(email string) error
	GetAll() ([]string, error)
}

type MailerConfig struct {
	Host                       string
	Port                       int
	Username, Password, Sender string
	UpdateInterval             time.Duration
}

func NewMailerService(
	cfg MailerConfig,
	emailService EmailService,
	rateService RateService,
) *mailer {
	dialer := mail.NewDialer(cfg.Host, cfg.Port, cfg.Username, cfg.Password)
	dialer.Timeout = DialerTimeout
	return &mailer{
		dialer:         dialer,
		sender:         cfg.Sender,
		emailService:   emailService,
		rateService:    rateService,
		updateInterval: cfg.UpdateInterval,
		stopChan:       make(chan bool),
	}
}

func (m *mailer) StartEmailSending() {
	ticker := time.NewTicker(m.updateInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				err := m.sendEmails()
				if err != nil {
					log.Error().Err(err).Send()
				}
			case <-m.stopChan:
				ticker.Stop()
				return
			}
		}
	}()
}

func (m *mailer) StopEmailSending() {
	m.stopChan <- true
}

func (m *mailer) sendEmails() error {
	message := mail.NewMessage()
	sender, err := m.dialer.Dial()
	if err != nil {
		return err
	}
	rate, err := m.rateService.GetRate(context.Background(), "usd")
	if err != nil {
		return err
	}
	err = m.renderCurrencyMessage(rate, message)
	if err != nil {
		return err
	}
	emails, err := m.emailService.GetAll()
	if err != nil {
		return err
	}

	var sendingError error
	for _, email := range emails {
		message.SetHeader("To", email)
		err = mail.Send(sender, message)
		if err != nil {
			sendingError = err
		}
		message.Reset()
	}
	if sendingError != nil {
		return sendingError
	}
	return sender.Close()
}

func (m *mailer) renderCurrencyMessage(data float32, message *mail.Message) error {
	tmpl, err := template.New("email").Parse(templates.MessageTemplate)
	if err != nil {
		return err
	}
	message.SetHeader("From", m.sender)

	headers := map[string]string{"subject": "Subject", "plainBody": "text/plain", "htmlBody": "text/html"}
	templateBuffer := new(bytes.Buffer)
	for templateName, header := range headers {
		err = tmpl.ExecuteTemplate(templateBuffer, templateName, data)
		if err != nil {
			return err
		}
		message.SetHeader(header, templateBuffer.String())
		templateBuffer.Reset()
	}

	return nil
}
