package logging

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/c653labs/pggateway"
	"github.com/rs/zerolog"
)

func init() {
	zerolog.TimeFieldFormat = ""
	pggateway.RegisterLoggingPlugin("file", newLoggingPlugin)
}

type LoggingPlugin struct {
	log zerolog.Logger
}

func newLoggingPlugin(config pggateway.ConfigMap) (pggateway.LoggingPlugin, error) {
	var err error

	var outFile io.Writer
	outFile = os.Stdout
	textColor := true
	out := config.StringDefault("out", "-")
	switch out {
	case "-":
		outFile = os.Stdout
		textColor = true
	default:
		outFile, err = os.OpenFile(out, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		textColor = false
		if err != nil {
			return nil, err
		}
	}

	format := strings.ToLower(config.StringDefault("format", "json"))
	if format != "text" && format != "json" {
		return nil, fmt.Errorf("unknown log format %#v, expected 'text' or 'json'", format)
	}

	if format == "text" {
		outFile = zerolog.ConsoleWriter{
			Out:     outFile,
			NoColor: !textColor,
		}
	}

	level := zerolog.WarnLevel
	l := config.StringDefault("level", "warn")
	l = strings.ToLower(l)
	level, err = zerolog.ParseLevel(l)
	if err != nil {
		return nil, err
	}

	return &LoggingPlugin{
		log: zerolog.New(outFile).Level(level).With().Timestamp().Logger(),
	}, nil
}

func (l *LoggingPlugin) logMsg(e *zerolog.Event, context pggateway.LoggingContext, msg string, args ...interface{}) {
	if !e.Enabled() {
		return
	}

	if context != nil {
		e = e.Fields(map[string]interface{}{"context": context})
	}
	e.Msgf(msg, args...)
}

func (l *LoggingPlugin) LogInfo(context pggateway.LoggingContext, msg string, args ...interface{}) {
	l.logMsg(l.log.Info(), context, msg, args...)
}

func (l *LoggingPlugin) LogError(context pggateway.LoggingContext, msg string, args ...interface{}) {
	l.logMsg(l.log.Error(), context, msg, args...)
}

func (l *LoggingPlugin) LogDebug(context pggateway.LoggingContext, msg string, args ...interface{}) {
	l.logMsg(l.log.Debug(), context, msg, args...)
}

func (l *LoggingPlugin) LogFatal(context pggateway.LoggingContext, msg string, args ...interface{}) {
	l.logMsg(l.log.Fatal(), context, msg, args...)
}

func (l *LoggingPlugin) LogWarn(context pggateway.LoggingContext, msg string, args ...interface{}) {
	l.logMsg(l.log.Warn(), context, msg, args...)
}
