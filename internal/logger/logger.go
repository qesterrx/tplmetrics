// Пакет logger - кастомный логгер основанный на github.com/rs/zerolog
package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// Log - Переменная содержащая ссылку на логгер, доступный из любого места проекта
var Log zerolog.Logger = zerolog.New(io.Discard)

// InitLogger - первоначальная настройка логгера, заданного в переменной Log
func InitLogger() {
	io := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	log := zerolog.New(io).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	Log = log
}
