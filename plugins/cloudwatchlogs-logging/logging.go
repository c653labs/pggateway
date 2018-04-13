package logging

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/c653labs/pggateway"
)

func init() {
	pggateway.RegisterLoggingPlugin("cloudwatchlogs", newLoggingPlugin)
}

type logLevel int

const (
	LevelFatal logLevel = iota
	LevelError
	LevelDebug
	LevelWarn
	LevelInfo
)

type LoggingPlugin struct {
	sess   *session.Session
	log    *cloudwatchlogs.CloudWatchLogs
	group  string
	stream string
	token  *string
	level  logLevel
}

func newLoggingPlugin(config map[string]string) (pggateway.LoggingPlugin, error) {
	options := session.Options{}
	region, ok := config["region"]
	if ok {
		options.Config = aws.Config{Region: aws.String(region)}
	}

	sess := session.Must(session.NewSessionWithOptions(options))
	logs := cloudwatchlogs.New(sess)

	// Log level
	level := LevelWarn
	if l, ok := config["level"]; ok {
		switch strings.ToLower(l) {
		case "warn":
			level = LevelWarn
		case "info":
			level = LevelInfo
		case "error":
			level = LevelError
		case "debug":
			level = LevelDebug
		case "fatal":
			level = LevelFatal
		default:
			return nil, fmt.Errorf("unknown logging level: %#v", level)
		}
	}

	group, ok := config["group"]
	if !ok {
		return nil, fmt.Errorf("must supply 'group' parameter")
	}

	stream, ok := config["stream"]
	if !ok {
		return nil, fmt.Errorf("must supply 'stream' parameter")
	}

	out, err := logs.DescribeLogStreams(&cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName:        aws.String(group),
		LogStreamNamePrefix: aws.String(stream),
	})
	if err != nil {
		return nil, err
	}

	exists := false
	var token *string
	for _, s := range out.LogStreams {
		if *s.LogStreamName == stream {
			token = s.UploadSequenceToken
			exists = true
			break
		}
	}

	if !exists {
		_, err = logs.CreateLogStream(&cloudwatchlogs.CreateLogStreamInput{
			LogGroupName:  aws.String(group),
			LogStreamName: aws.String(stream),
		})
		if err != nil {
			return nil, err
		}
	}

	return &LoggingPlugin{
		sess:   sess,
		log:    logs,
		level:  level,
		group:  group,
		stream: stream,
		token:  token,
	}, nil
}

func (l *LoggingPlugin) putLogEvent(level logLevel, context pggateway.LoggingContext, msg string, args ...interface{}) error {
	if level < l.level {
		return nil
	}

	now := aws.TimeUnixMilli(time.Now())
	msgMap := map[string]interface{}{
		"context":   context,
		"timestamp": now,
		"text":      fmt.Sprintf(msg, args...),
	}

	msgFormatted, err := json.Marshal(msgMap)
	if err != nil {
		return err
	}

	res, err := l.log.PutLogEvents(&cloudwatchlogs.PutLogEventsInput{
		LogGroupName:  aws.String(l.group),
		LogStreamName: aws.String(l.stream),
		SequenceToken: l.token,
		LogEvents: []*cloudwatchlogs.InputLogEvent{
			&cloudwatchlogs.InputLogEvent{
				Message:   aws.String(string(msgFormatted)),
				Timestamp: aws.Int64(now),
			},
		},
	})
	if err != nil {
		return err
	}

	if res.RejectedLogEventsInfo != nil {
		return fmt.Errorf("rejected log events info: %#v", res.RejectedLogEventsInfo)
	}

	l.token = res.NextSequenceToken
	return nil
}

func (l *LoggingPlugin) LogInfo(context pggateway.LoggingContext, msg string, args ...interface{}) {
	err := l.putLogEvent(LevelInfo, context, msg, args...)
	if err != nil {
		log.Println(err)
	}
}

func (l *LoggingPlugin) LogError(context pggateway.LoggingContext, msg string, args ...interface{}) {
	l.putLogEvent(LevelError, context, msg, args...)
}

func (l *LoggingPlugin) LogDebug(context pggateway.LoggingContext, msg string, args ...interface{}) {
	l.putLogEvent(LevelDebug, context, msg, args...)
}

func (l *LoggingPlugin) LogFatal(context pggateway.LoggingContext, msg string, args ...interface{}) {
	l.putLogEvent(LevelFatal, context, msg, args...)
}

func (l *LoggingPlugin) LogWarn(context pggateway.LoggingContext, msg string, args ...interface{}) {
	l.putLogEvent(LevelWarn, context, msg, args...)
}
