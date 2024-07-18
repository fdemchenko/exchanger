package config

import (
	"flag"
	"os"
	"time"
)

type Config struct {
	SMTP               SMTPConfig
	RabbitMQConnString string
	SchedulerInterval  time.Duration
}

type SMTPConfig struct {
	Host               string
	Username           string
	Port               int
	Password           string
	Sender             string
	ConnectionPoolSize int
}

const (
	DefaultSMTPPort                 = 25
	DefaultRabbitMQPort             = 5672
	DefaultMailerConnectionPoolSize = 3
	DefaultSchedulerInterval        = 24 * time.Hour
)

func LoadConfig() Config {
	var cfg Config
	flag.StringVar(&cfg.SMTP.Host, "smtp-host", os.Getenv("EXCHANGER_SMTP_HOST"), "Smtp host")
	flag.IntVar(&cfg.SMTP.Port, "smtp-port", DefaultSMTPPort, "Smtp port")
	flag.IntVar(&cfg.SMTP.ConnectionPoolSize,
		"smtp-connections",
		DefaultMailerConnectionPoolSize,
		"Smtp connection pool size",
	)
	flag.Func("scheduler-interval", "Email update interval (E.g. 24h, 1h30m)", func(s string) error {
		if s == "" {
			cfg.SchedulerInterval = DefaultSchedulerInterval
			return nil
		}
		duration, err := time.ParseDuration(s)
		if err != nil {
			return err
		}
		cfg.SchedulerInterval = duration
		return nil
	})
	flag.StringVar(&cfg.SMTP.Username, "smtp-username", os.Getenv("EXCHANGER_SMTP_USERNAME"), "Smtp username")
	flag.StringVar(&cfg.SMTP.Password, "smtp-password", os.Getenv("EXCHANGER_SMTP_PASSWORD"), "Smtp password")
	flag.StringVar(&cfg.SMTP.Sender, "smtp-sender", os.Getenv("EXCHANGER_SMTP_SENDER"), "Smtp sender")

	flag.StringVar(&cfg.RabbitMQConnString,
		"rabbitmq-conn-string",
		os.Getenv("EXCHANGER_RABBITMQ_CONN_STRING"),
		"RabbitMQ connection string",
	)
	flag.Parse()
	return cfg
}
