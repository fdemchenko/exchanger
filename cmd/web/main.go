package main

import (
	"database/sql"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/fdemchenko/exchanger/internal/models"
	"github.com/fdemchenko/exchanger/internal/services"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

type config struct {
	addr string
	db   struct {
		dsn            string
		maxConnections int
	}
	smtp services.MailerConfig
}

type application struct {
	cfg               config
	rateService       *services.RateService
	errorLog, infoLog *log.Logger
	emailModel        *models.EmailModel
}

func main() {
	var cfg config
	flag.StringVar(&cfg.addr, "addr", ":8080", "http listen address")
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("EXCHANGER_DSN"), "Data source name")
	flag.IntVar(&cfg.db.maxConnections, "db-max-conn", 25, "Database max connection")

	flag.StringVar(&cfg.smtp.Host, "smtp-host", os.Getenv("EXCHANGER_SMPT_HOST"), "Smpt host")
	flag.IntVar(&cfg.smtp.Port, "smtp-port", 25, "Smpt port")
	flag.StringVar(&cfg.smtp.Username, "smtp-username", os.Getenv("EXCHANGER_SMPT_USERNAME"), "Smpt username")
	flag.StringVar(&cfg.smtp.Password, "smtp-password", os.Getenv("EXCHANGER_SMPT_PASSWORD"), "Smpt password")
	flag.StringVar(&cfg.smtp.Sender, "smtp-sender", os.Getenv("EXCHANGER_SMPT_SENDER"), "Smpt sender")
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO ", log.Ldate|log.Lshortfile)
	errorLog := log.New(os.Stderr, "ERROR ", log.Ldate|log.Lshortfile)

	db, err := openDB(cfg)
	if err != nil {
		errorLog.Fatalln(err)
	}
	infoLog.Println("Coonected to DB successfuly")

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		errorLog.Fatalln(err)
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres", driver)
	if err != nil {
		errorLog.Fatalln(err)
	}
	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		errorLog.Fatalln(err)
	}

	emailModel := &models.EmailModel{DB: db}
	rateService := services.NewRateService(time.Hour)

	mailerService := services.NewMailerService(cfg.smtp, emailModel, rateService, errorLog)
	mailerService.StartBackgroundTask(time.Hour * 24)

	app := application{
		cfg:         cfg,
		rateService: rateService,
		errorLog:    errorLog,
		infoLog:     infoLog,
		emailModel:  emailModel,
	}

	server := http.Server{
		Handler:           app.routes(),
		Addr:              app.cfg.addr,
		WriteTimeout:      10 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
	}

	infoLog.Println("Starting web server at " + app.cfg.addr)
	errorLog.Fatalln(server.ListenAndServe())
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
