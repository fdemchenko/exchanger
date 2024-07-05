package main

import (
	"github.com/rs/zerolog/log"
)

const (
	DefaultSMTPPort                 = 25
	DefaultRabbitMQPort             = 5672
	DefaultMailerConnectionPoolSize = 3
)

func main() {
	log.Info().Msg("Mialer service entrypoint")
}
