package services

import (
	"bytes"
	"log"
	"text/template"
	"time"

	_ "embed"

	"github.com/fdemchenko/exchanger/internal/models"
	"github.com/go-mail/mail/v2"
)

//go:embed "rate_update.tmpl"
var messageTemplate string

type Mailer struct {
	dialer      *mail.Dialer
	sender      string
	emailModel  *models.EmailModel
	rateService *RateService
	errorLog    *log.Logger
}

type MailerConfig struct {
	Host                       string
	Port                       int
	Username, Password, Sender string
}

func NewMailerService(cfg MailerConfig, emailModel *models.EmailModel, rateService *RateService, errorLog *log.Logger) Mailer {
	dialer := mail.NewDialer(cfg.Host, cfg.Port, cfg.Username, cfg.Password)
	dialer.Timeout = 5 * time.Second
	return Mailer{
		dialer:      dialer,
		sender:      cfg.Sender,
		emailModel:  emailModel,
		rateService: rateService,
		errorLog:    errorLog,
	}
}

func (m Mailer) StartBackgroundTask(interval time.Duration) {
	go func() {
		for range time.Tick(interval) {
			rate, err := m.rateService.GetRate()
			if err != nil {
				m.errorLog.Println(err)
				continue
			}
			msg, err := m.prepareMessage(rate)
			if err != nil {
				m.errorLog.Println(err)
				continue
			}
			emails, err := m.emailModel.GetAll()
			if err != nil {
				m.errorLog.Println(err)
				continue
			}

			for _, email := range emails {
				msg.SetHeader("To", email)
				err = m.dialer.DialAndSend(msg)
				if err != nil {
					m.errorLog.Println(err)
				}
			}
		}
	}()
}

func (m Mailer) prepareMessage(data interface{}) (*mail.Message, error) {
	tmpl, err := template.New("email").Parse(messageTemplate)
	if err != nil {
		return nil, err
	}

	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return nil, err
	}

	plainBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(plainBody, "plainBody", data)
	if err != nil {
		return nil, err
	}

	htmlBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(htmlBody, "htmlBody", data)
	if err != nil {
		return nil, err
	}

	msg := mail.NewMessage()
	msg.SetHeader("From", m.sender)
	msg.SetHeader("Subject", subject.String())
	msg.SetBody("text/plain", plainBody.String())
	msg.AddAlternative("text/html", htmlBody.String())

	return msg, nil
}
