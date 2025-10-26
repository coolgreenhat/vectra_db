package logger

import (
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

type Logger struct {
	*logrus.Logger
}

type Config struct {
	Level  string
	Format string
}

func New(config Config) *Logger {
	log := logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	log.SetLevel(level)

	// Set log format
	if config.Format == "json" {
		log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
	} else {
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339,
		})
	}

	// Set output
	log.SetOutput(os.Stdout)

	return &Logger{Logger: log}
}

func (l *Logger) WithField(key string, value interface{}) *logrus.Entry {
	return l.Logger.WithField(key, value)
}

func (l *Logger) WithFields(fields logrus.Fields) *logrus.Entry {
	return l.Logger.WithFields(fields)
}

func (l *Logger) WithError(err error) *logrus.Entry {
	return l.Logger.WithError(err)
}

var Default *Logger

func Init(config Config) {
	Default = New(config)
}

func WithField(key string, value interface{}) *logrus.Entry {
	if Default == nil {
		Init(Config{Level: "info", Format: "json"})
	}
	return Default.WithField(key, value)
}

func WithFields(fields logrus.Fields) *logrus.Entry {
	if Default == nil {
		Init(Config{Level: "info", Format: "json"})
	}
	return Default.WithFields(fields)
}

func WithError(err error) *logrus.Entry {
	if Default == nil {
		Init(Config{Level: "info", Format: "json"})
	}
	return Default.WithError(err)
}

func Info(args ...interface{}) {
	if Default == nil {
		Init(Config{Level: "info", Format: "json"})
	}
	Default.Info(args...)
}

func Infof(format string, args ...interface{}) {
	if Default == nil {
		Init(Config{Level: "info", Format: "json"})
	}
	Default.Infof(format, args...)
}

func Error(args ...interface{}) {
	if Default == nil {
		Init(Config{Level: "info", Format: "json"})
	}
	Default.Error(args...)
}

func Errorf(format string, args ...interface{}) {
	if Default == nil {
		Init(Config{Level: "info", Format: "json"})
	}
	Default.Errorf(format, args...)
}

func Debug(args ...interface{}) {
	if Default == nil {
		Init(Config{Level: "info", Format: "json"})
	}
	Default.Debug(args...)
}

func Debugf(format string, args ...interface{}) {
	if Default == nil {
		Init(Config{Level: "info", Format: "json"})
	}
	Default.Debugf(format, args...)
}

func Warn(args ...interface{}) {
	if Default == nil {
		Init(Config{Level: "info", Format: "json"})
	}
	Default.Warn(args...)
}

func Warnf(format string, args ...interface{}) {
	if Default == nil {
		Init(Config{Level: "info", Format: "json"})
	}
	Default.Warnf(format, args...)
}

func Fatal(args ...interface{}) {
	if Default == nil {
		Init(Config{Level: "info", Format: "json"})
	}
	Default.Fatal(args...)
}

func Fatalf(format string, args ...interface{}) {
	if Default == nil {
		Init(Config{Level: "info", Format: "json"})
	}
	Default.Fatalf(format, args...)
}
