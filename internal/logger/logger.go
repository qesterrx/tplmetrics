package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

var Log zerolog.Logger = zerolog.New(io.Discard)

// Singleton
func InitLogger() {
	io := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	log := zerolog.New(io).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	Log = log
}
