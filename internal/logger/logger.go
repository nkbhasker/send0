package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/usesend0/send0/internal/config"
	"github.com/usesend0/send0/internal/constant"
)

func NewLogger(cnf *config.Config) zerolog.Logger {
	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}

	logLevel := func() zerolog.Level {
		if cnf.Env == constant.EnvDevelopment {
			return zerolog.DebugLevel
		}
		return zerolog.InfoLevel
	}

	logger := zerolog.New(output).
		Level(zerolog.Level(logLevel())).
		With().
		Timestamp().
		Logger()

	return logger
}
