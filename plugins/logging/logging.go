package logging

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/c653labs/pggateway"
)

func init() {
	pggateway.RegisterPlugin("logging", newLoggingPlugin())
}

type LoggingPlugin struct {
	log *logrus.Logger
}

func newLoggingPlugin() *LoggingPlugin {
	log := logrus.New()
	log.Formatter = &logrus.JSONFormatter{}
	log.Out = os.Stdout
	log.Level = logrus.InfoLevel

	return &LoggingPlugin{
		log: log,
	}
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
