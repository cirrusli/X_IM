package logger

import (
	"github.com/sirupsen/logrus"
)

var (
	std *logrus.Logger = logrus.New()
)

type Fields map[string]interface{}

func WithFields(fields Fields) *logrus.Entry {
	fie := logrus.Fields(fields)
	return std.WithFields(fie)
}
func Tracef(format string, args ...interface{}) {
	std.Tracef(format, args...)
}
func Error(args ...interface{}) {
	std.Error(args...)
}
