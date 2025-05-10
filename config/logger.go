package config

import (
	"github.com/sirupsen/logrus"
)

// Logger is the global logger instance used throughout the application.
// It is initialized with a default formatter and log level.
var Logger = logrus.New()

func init() {
	Logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	Logger.SetLevel(logrus.ErrorLevel)
}

// SetLogLevel allows overriding the log level dynamically.
//
// Parameters:
// - level: The desired log level (e.g., logrus.InfoLevel, logrus.ErrorLevel).
func SetLogLevel(level logrus.Level) {
	Logger.SetLevel(level)
}
