package logger

import (
	"context"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"path/filepath"
	"time"
)

var (
	std = logrus.New()
)

type Fields map[string]any

type Option func(opts *Options)

type Options struct {
	Format        string
	RotationCount int
	LogLevel      string
}

//func init() {
// 经常直接调用全局的logger.Info()，导致日志输出的文件名和行号不正确
//	std.SetReportCaller(true)
//}

// WithRotationCount set rotation count of log files
func WithRotationCount(v int) Option {
	return func(opts *Options) {
		opts.RotationCount = v
	}
}

// WithFormat set format of log files, text or json
func WithFormat(format string) Option {
	return func(opts *Options) {
		opts.Format = format
	}
}

// WithLevel set log level
func WithLevel(level string) Option {
	return func(opts *Options) {
		opts.LogLevel = level
	}
}

// InitDailyRolling init a logger with default 7 days remained
func InitDailyRolling(fileDir, fileName string, opts ...Option) error {
	logfile := filepath.Join(fileDir, fileName)
	options := &Options{
		Format:        "text",
		RotationCount: 7,
		LogLevel:      "debug",
	}
	for _, opt := range opts {
		opt(options)
	}

	writer, err := rotatelogs.New(
		logfile+".%Y%m%d",
		// WithLinkName为最新的日志建立软连接，以方便随着找到当前日志文件
		rotatelogs.WithLinkName(logfile),

		// WithRotationTime设置日志分割的时间，这里设置为一小时分割一次
		rotatelogs.WithRotationTime(time.Hour*24),

		// WithMaxAge和WithRotationCount二者只能设置一个，
		// WithMaxAge设置文件清理前的最长保存时间，
		// WithRotationCount设置文件清理前最多保存的个数。
		//rotatelogs.WithMaxAge(time.Hour*24),
		rotatelogs.WithRotationCount(uint(options.RotationCount)),
	)
	if err != nil {
		return err
	}

	var logfr logrus.Formatter
	if options.Format == "json" {
		logfr = &logrus.JSONFormatter{
			DisableTimestamp: false,
		}
	} else {
		logfr = &logrus.TextFormatter{DisableColors: true}
	}
	_ = SetLevel(options.LogLevel)

	lfsHook := lfshook.NewHook(lfshook.WriterMap{
		logrus.InfoLevel:  writer,
		logrus.WarnLevel:  writer,
		logrus.ErrorLevel: writer,
		logrus.FatalLevel: writer,
		logrus.PanicLevel: writer,
	}, logfr)
	std.AddHook(lfsHook)

	std.Info("***********This is logrus*************")
	return nil
}

// SetLevel SetLevel
func SetLevel(level string) error {
	ll, err := logrus.ParseLevel(level)
	if err == nil {
		std.SetLevel(ll)
	}
	return err
}

type Entry *logrus.Entry

// WithError creates an entry from the standard logger and adds an error to it, using the value defined in ErrorKey as key.
func WithError(err error) *logrus.Entry {
	return std.WithField(logrus.ErrorKey, err)
}

// WithContext creates an entry from the standard logger and adds a context to it.
func WithContext(ctx context.Context) *logrus.Entry {
	return std.WithContext(ctx)
}

// WithField creates an entry from the standard logger and adds a field to
// it. If you want multiple fields, use `WithFields`.
//
// Note that it doesn't log until you call Debug, Print, Info, Warn, Fatal
// or Panic on the  *logrus.Entry  it returns.
func WithField(key string, value any) *logrus.Entry {
	return std.WithField(key, value)
}

// WithFields creates an entry from the standard logger and adds multiple
// fields to it. This is simply a helper for `WithField`, invoking it
// once for each field.
//
// Note that it doesn't log until you call Debug, Print, Info, Warn, Fatal
// or Panic on the  *logrus.Entry  it returns.
func WithFields(fields Fields) *logrus.Entry {
	fie := logrus.Fields(fields)
	return std.WithFields(fie)
}

// WithTime creates an entry from the standard logger and overrides the time of
// logs generated with it.
//
// Note that it doesn't log until you call Debug, Print, Info, Warn, Fatal
// or Panic on the  *logrus.Entry  it returns.
func WithTime(t time.Time) *logrus.Entry {
	return std.WithTime(t)
}

// Trace logs a message at level Trace on the standard logger.
func Trace(args ...any) {
	std.Trace(args...)
}

// Debug logs a message at level Debug on the standard logger.
func Debug(args ...any) {
	std.Debug(args...)
}

// Info logs a message at level Info on the standard logger.
func Info(args ...any) {
	std.Info(args...)
}

// Warn logs a message at level Warn on the standard logger.
func Warn(args ...any) {
	std.Warn(args...)
}

// Error logs a message at level Error on the standard logger.
func Error(args ...any) {
	std.Error(args...)
}

// Panic logs a message at level Panic on the standard logger.
func Panic(args ...any) {
	std.Panic(args...)
}

// Fatal logs a message at level Fatal on the standard logger then the process will exit with status set to 1.
func Fatal(args ...any) {
	std.Fatal(args...)
}

// Tracef logs a message at level Trace on the standard logger.
func Tracef(format string, args ...any) {
	std.Tracef(format, args...)
}

// Debugf logs a message at level Debug on the standard logger.
func Debugf(format string, args ...any) {
	std.Debugf(format, args...)
}

// Printf logs a message at level Info on the standard logger.
func Printf(format string, args ...any) {
	std.Printf(format, args...)
}

// Infof logs a message at level Info on the standard logger.
func Infof(format string, args ...any) {
	std.Infof(format, args...)
}

// Warnf logs a message at level Warn on the standard logger.
func Warnf(format string, args ...any) {
	std.Warnf(format, args...)
}

// Errorf logs a message at level Error on the standard logger.
func Errorf(format string, args ...any) {
	std.Errorf(format, args...)
}

// Panicf logs a message at level Panic on the standard logger.
func Panicf(format string, args ...any) {
	std.Panicf(format, args...)
}

// Fatalf logs a message at level Fatal on the standard logger then the process will exit with status set to 1.
func Fatalf(format string, args ...any) {
	std.Fatalf(format, args...)
}

// Traceln logs a message at level Trace on the standard logger.
func Traceln(args ...any) {
	std.Traceln(args...)
}

// Debugln logs a message at level Debug on the standard logger.
func Debugln(args ...any) {
	std.Debugln(args...)
}

// Println logs a message at level Info on the standard logger.
func Println(args ...any) {
	std.Println(args...)
}

// Infoln logs a message at level Info on the standard logger.
func Infoln(args ...any) {
	std.Infoln(args...)
}

// Warnln logs a message at level Warn on the standard logger.
func Warnln(args ...any) {
	std.Warnln(args...)
}

// Warningln logs a message at level Warn on the standard logger.
func Warningln(args ...any) {
	std.Warningln(args...)
}

// Errorln logs a message at level Error on the standard logger.
func Errorln(args ...any) {
	std.Errorln(args...)
}

// Panicln logs a message at level Panic on the standard logger.
func Panicln(args ...any) {
	std.Panicln(args...)
}

// Fatalln logs a message at level Fatal on the standard logger then the process will exit with status set to 1.
func Fatalln(args ...any) {
	std.Fatalln(args...)
}
