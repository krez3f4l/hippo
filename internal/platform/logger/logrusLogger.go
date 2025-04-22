package logger

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"

	"hippo/internal/platform/consts"
)

type logrusLogger struct {
	entry *logrus.Entry
}

func SetupLogrusLogger(env string) Logger {
	base := logrus.New()
	base.SetOutput(os.Stdout)

	switch {
	case strings.EqualFold(env, consts.EnvLocal):
		base.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
		base.SetLevel(logrus.DebugLevel)

	case strings.EqualFold(env, consts.EnvDev):
		base.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05Z07:00",
		})
		base.SetLevel(logrus.InfoLevel)

	case strings.EqualFold(env, consts.EnvProd):
		base.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05Z07:00",
		})
		base.SetLevel(logrus.InfoLevel)
	}

	return &logrusLogger{
		entry: logrus.NewEntry(base),
	}
}

func (l *logrusLogger) Debug(msg string, fields ...Field) {
	logFields := make(logrus.Fields)
	for _, field := range fields {
		logFields[field.Key] = field.Value
	}
	l.entry.WithFields(logFields).Debug(msg)
}

func (l *logrusLogger) Info(msg string, fields ...Field) {
	logFields := make(logrus.Fields)
	for _, field := range fields {
		logFields[field.Key] = field.Value
	}
	l.entry.WithFields(logFields).Info(msg)
}

func (l *logrusLogger) Warn(msg string, fields ...Field) {
	logFields := make(logrus.Fields)
	for _, field := range fields {
		logFields[field.Key] = field.Value
	}
	l.entry.WithFields(logFields).Warn(msg)
}

func (l *logrusLogger) Error(msg string, fields ...Field) {
	logFields := make(logrus.Fields)
	for _, field := range fields {
		logFields[field.Key] = field.Value
	}
	l.entry.WithFields(logFields).Error(msg)
}

func (l *logrusLogger) Fatal(msg string, fields ...Field) {
	logFields := make(logrus.Fields)
	for _, field := range fields {
		logFields[field.Key] = field.Value
	}
	l.entry.WithFields(logFields).Fatal(msg)
}

func (l *logrusLogger) WithFields(fields map[string]any) Logger {
	return &logrusLogger{
		entry: l.entry.WithFields(fields),
	}
}
