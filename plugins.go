package pggateway

import (
	"fmt"
	"sync"
)

var authPlugins = make(map[string]authPluginInitializer)
var loggingPlugins = make(map[string]loggingPluginInitializer)

type authPluginInitializer func(interface{}) (AuthenticationPlugin, error)
type loggingPluginInitializer func(ConfigMap) (LoggingPlugin, error)

type Plugin interface{}

type AuthenticationPlugin interface {
	Plugin
	Authenticate(*Session) (bool, error)
}

type LoggingContext map[string]interface{}

type LoggingPlugin interface {
	Plugin
	LogInfo(LoggingContext, string, ...interface{})
	LogDebug(LoggingContext, string, ...interface{})
	LogError(LoggingContext, string, ...interface{})
	LogFatal(LoggingContext, string, ...interface{})
	LogWarn(LoggingContext, string, ...interface{})
}

func RegisterAuthPlugin(name string, init authPluginInitializer) {
	authPlugins[name] = init
}

func RegisterLoggingPlugin(name string, init func(ConfigMap) (LoggingPlugin, error)) {
	loggingPlugins[name] = init
}

type loggingMessage struct {
	level   string
	context LoggingContext
	msg     string
	args    []interface{}
}

type PluginRegistry struct {
	authPlugins    map[string]AuthenticationPlugin
	loggingPlugins map[string]LoggingPlugin
	logMutex       *sync.Mutex
}

func NewPluginRegistry(auth map[string]interface{}, logging map[string]ConfigMap) (*PluginRegistry, error) {
	r := &PluginRegistry{
		authPlugins:    make(map[string]AuthenticationPlugin),
		loggingPlugins: make(map[string]LoggingPlugin),
		logMutex:       &sync.Mutex{},
	}

	for name, config := range auth {
		init, ok := authPlugins[name]
		if !ok {
			return nil, fmt.Errorf("could not find authentication plugin: %s", name)
		}

		p, err := init(config)
		if err != nil {
			return nil, err
		}
		r.authPlugins[name] = p
	}

	for name, config := range logging {
		init, ok := loggingPlugins[name]
		if !ok {
			return nil, fmt.Errorf("could not find logging plugin: %s", name)
		}

		p, err := init(config)
		if err != nil {
			return nil, err
		}
		r.loggingPlugins[name] = p
	}

	return r, nil
}

func (r *PluginRegistry) handleLog(msg loggingMessage) {
	r.logMutex.Lock()
	for _, p := range r.loggingPlugins {
		switch msg.level {
		case "info":
			p.LogInfo(msg.context, msg.msg, msg.args...)
		case "debug":
			p.LogDebug(msg.context, msg.msg, msg.args...)
		case "warn":
			p.LogWarn(msg.context, msg.msg, msg.args...)
		case "error":
			p.LogError(msg.context, msg.msg, msg.args...)
		case "fatal":
			p.LogFatal(msg.context, msg.msg, msg.args...)
		}
	}
	r.logMutex.Unlock()
}

func (r *PluginRegistry) Authenticate(sess *Session) (bool, error) {
	for _, p := range r.authPlugins {
		success, err := p.Authenticate(sess)
		if err != nil {
			return false, err
		}
		if success {
			return true, nil
		}
	}

	return false, nil
}

func (r *PluginRegistry) LogInfo(context LoggingContext, msg string, args ...interface{}) {
	r.handleLog(loggingMessage{
		level:   "info",
		context: context,
		msg:     msg,
		args:    args,
	})
}

func (r *PluginRegistry) LogError(context LoggingContext, msg string, args ...interface{}) {
	r.handleLog(loggingMessage{
		level:   "error",
		context: context,
		msg:     msg,
		args:    args,
	})
}

func (r *PluginRegistry) LogWarn(context LoggingContext, msg string, args ...interface{}) {
	r.handleLog(loggingMessage{
		level:   "warn",
		context: context,
		msg:     msg,
		args:    args,
	})
}

func (r *PluginRegistry) LogDebug(context LoggingContext, msg string, args ...interface{}) {
	r.handleLog(loggingMessage{
		level:   "debug",
		context: context,
		msg:     msg,
		args:    args,
	})
}

func (r *PluginRegistry) LogFatal(context LoggingContext, msg string, args ...interface{}) {
	r.handleLog(loggingMessage{
		level:   "fatal",
		context: context,
		msg:     msg,
		args:    args,
	})
}
