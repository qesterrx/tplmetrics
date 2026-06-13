package logger

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

func TestInitLogger(t *testing.T) {
	// Сохраняем оригинальный логгер
	oldLog := Log
	defer func() {
		Log = oldLog
	}()

	InitLogger()

	// Проверяем, что логгер инициализирован
	if Log.GetLevel() == zerolog.Disabled {
		t.Error("Logger should be initialized")
	}

	// Проверяем уровень логирования
	if zerolog.GlobalLevel() != zerolog.InfoLevel {
		t.Errorf("Expected InfoLevel, got %v", zerolog.GlobalLevel())
	}
}

func TestLogMessages(t *testing.T) {
	// Сохраняем оригинальный логгер
	oldLog := Log
	defer func() {
		Log = oldLog
	}()

	// Создаем буфер для перехвата вывода
	var buf bytes.Buffer
	writer := zerolog.ConsoleWriter{Out: &buf, TimeFormat: time.RFC3339, NoColor: true}
	Log = zerolog.New(writer).With().Timestamp().Logger()

	tests := []struct {
		name     string
		logFunc  func()
		expected string
	}{
		{
			name: "Info message",
			logFunc: func() {
				Log.Info().Msg("test info")
			},
			expected: "test info",
		},
		{
			name: "Error message",
			logFunc: func() {
				Log.Error().Msg("test error")
			},
			expected: "test error",
		},
		{
			name: "Warn message",
			logFunc: func() {
				Log.Warn().Msg("test warn")
			},
			expected: "test warn",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.logFunc()

			if !strings.Contains(buf.String(), tt.expected) {
				t.Errorf("Expected log to contain '%s', got '%s'", tt.expected, buf.String())
			}
		})
	}
}
