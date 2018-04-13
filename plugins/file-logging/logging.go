package logging

import (
	"fmt"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/c653labs/pggateway"
)

func init() {
	pggateway.RegisterLoggingPlugin("file", newLoggingPlugin)
}

type LoggingPlugin struct {
	log *logrus.Logger
}

func newLoggingPlugin(config map[string]string) (pggateway.LoggingPlugin, error) {
	log := logrus.New()

	// Log format
	log.Formatter = &logrus.TextFormatter{
		FullTimestamp:    true,
		DisableTimestamp: false,
	}
	if format, ok := config["format"]; ok {
		switch strings.ToLower(format) {
		case "json":
			log.Formatter = &logrus.JSONFormatter{
				DisableTimestamp: false,
				FieldMap: logrus.FieldMap{
					logrus.FieldKeyMsg: "text",
				},
			}
		}
	}

	// Out file
	log.Out = os.Stdout
	if out, ok := config["out"]; ok {
		switch out {
		case "-":
			log.Out = os.Stdout
		}
	}

	// Log level
	log.Level = logrus.WarnLevel
	if level, ok := config["level"]; ok {
		switch strings.ToLower(level) {
		case "warn":
			log.Level = logrus.WarnLevel
		case "info":
			log.Level = logrus.InfoLevel
		case "error":
			log.Level = logrus.ErrorLevel
		case "debug":
			log.Level = logrus.DebugLevel
		case "fatal":
			log.Level = logrus.FatalLevel
		default:
			return nil, fmt.Errorf("unknown logging level: %#v", level)
		}
	}

	return &LoggingPlugin{
		log: log,
	}, nil
}

func (l *LoggingPlugin) entry(context pggateway.LoggingContext) *logrus.Entry {
	entry := logrus.NewEntry(l.log)
	if context != nil {
		entry = entry.WithFields((logrus.Fields)(context))
	}
	return entry
}

func (l *LoggingPlugin) LogInfo(context pggateway.LoggingContext, msg string, args ...interface{}) {
	l.entry(context).Infof(msg, args...)
}

func (l *LoggingPlugin) LogError(context pggateway.LoggingContext, msg string, args ...interface{}) {
	l.entry(context).Errorf(msg, args...)
}

func (l *LoggingPlugin) LogDebug(context pggateway.LoggingContext, msg string, args ...interface{}) {
	l.entry(context).Debugf(msg, args...)
}

func (l *LoggingPlugin) LogFatal(context pggateway.LoggingContext, msg string, args ...interface{}) {
	l.entry(context).Warnf(msg, args...)
}

func (l *LoggingPlugin) LogWarn(context pggateway.LoggingContext, msg string, args ...interface{}) {
	l.entry(context).Warnf(msg, args...)
}
