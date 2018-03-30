package logging

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/c653labs/pggateway"
)

func init() {
	pggateway.RegisterLoggingPlugin("logging", newLoggingPlugin)
}

type LoggingPlugin struct {
	log *logrus.Logger
}

func newLoggingPlugin() (pggateway.LoggingPlugin, error) {
	log := logrus.New()
	log.Formatter = &logrus.TextFormatter{
		FullTimestamp:    true,
		DisableTimestamp: false,
	}
	log.Out = os.Stdout
	log.Level = logrus.WarnLevel

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
