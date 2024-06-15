package main

import (
	"database/sql"
	"errors"
	"flag"
	"net/http"
	"os"
	"time"

	"github.com/fdemchenko/exchanger/internal/repositories"
	"github.com/fdemchenko/exchanger/internal/services"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type config struct {
	addr string
	db   struct {
		dsn            string
		maxConnections int
	}
	mailer services.MailerConfig
}

type application struct {
	cfg          config
	rateService  services.RateService
	emailService services.EmailService
}

const (
	DefaultSMTPPort         = 25
	ServerTimeout           = 10 * time.Second
	DefaultMaxDBConnections = 25
	DefaultMailerInterval   = 24 * time.Hour
)

func main() {
	var cfg config
	flag.StringVar(&cfg.addr, "addr", ":8080", "http listen address")
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("EXCHANGER_DSN"), "Data source name")
	flag.IntVar(&cfg.db.maxConnections, "db-max-conn", DefaultMaxDBConnections, "Database max connection")

	flag.StringVar(&cfg.mailer.Host, "smtp-host", os.Getenv("EXCHANGER_SMPT_HOST"), "Smpt host")
	flag.IntVar(&cfg.mailer.Port, "smtp-port", DefaultSMTPPort, "Smpt port")
	flag.StringVar(&cfg.mailer.Username, "smtp-username", os.Getenv("EXCHANGER_SMPT_USERNAME"), "Smpt username")
	flag.StringVar(&cfg.mailer.Password, "smtp-password", os.Getenv("EXCHANGER_SMPT_PASSWORD"), "Smpt password")
	flag.StringVar(&cfg.mailer.Sender, "smtp-sender", os.Getenv("EXCHANGER_SMPT_SENDER"), "Smpt sender")
	flag.Func("mailer-interval", "Email update interval (E.g. 24h, 1h30m)", func(s string) error {
		if s == "" {
			cfg.mailer.UpdateInterval = DefaultMailerInterval
			return nil
		}
		duration, err := time.ParseDuration(s)
		if err != nil {
			return err
		}
		cfg.mailer.UpdateInterval = duration
		return nil
	})
	flag.Parse()

	zerolog.TimeFieldFormat = time.RFC3339

	db, err := openDB(cfg)
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	log.Print("Coonected to DB successfully")

	err = autoMigrate(db)
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	emailRepository := &repositories.PostgresEmailRepository{DB: db}
	emailService := services.NewEmailService(emailRepository)
	rateService := services.NewRateService(time.Hour)
	rateService.StartBackgroundTask()

	mailerService := services.NewMailerService(cfg.mailer, emailService, rateService)
	mailerService.StartBackgroundTask()

	app := application{
		cfg:          cfg,
		rateService:  rateService,
		emailService: emailService,
	}

	server := http.Server{
		Handler:           app.routes(),
		Addr:              app.cfg.addr,
		WriteTimeout:      ServerTimeout,
		ReadHeaderTimeout: ServerTimeout,
	}

	log.Info().Msg("Starting web server at " + app.cfg.addr)
	log.Fatal().Err(server.ListenAndServe()).Send()
}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(cfg.db.maxConnections)
	return db, nil
}

func autoMigrate(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres", driver)
	if err != nil {
		return err
	}
	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}
