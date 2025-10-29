package logging

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"time"
)

func InitializeLogging() {
	zerolog.TimeFieldFormat = time.RFC3339
	log.Logger = log.Output(os.Stdout)
	// TODO integrate with go-chi HTTP server
}
