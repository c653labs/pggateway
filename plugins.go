package pggateway

import (
	"github.com/c653labs/pgproto"
)

var authPlugin AuthenticationPlugin
var loggingPlugins = make(map[string]LoggingPlugin)

type Plugin interface{}

type AuthenticationPlugin interface {
	Plugin
	OnStart()
	OnValidate()
	OnFinalize()
}

type LoggingPlugin interface {
	Plugin
	LogSystem(string, ...interface{})
	LogNewSession(*Session)
	LogClientRequest(*Session, pgproto.ClientMessage)
	LogServerResponse(*Session, pgproto.ServerMessage)
	LogSessionClosed(*Session, error)
}

func RegisterPlugin(name string, plugin Plugin) {
	switch p := plugin.(type) {
	case AuthenticationPlugin:
		authPlugin = p
	case LoggingPlugin:
		loggingPlugins[name] = p
	}
}

type PluginRegistry struct{}

func NewPluginRegistry() PluginRegistry { return PluginRegistry{} }

func (r PluginRegistry) LogSystem(fmt string, args ...interface{}) {
	for _, p := range loggingPlugins {
		p.LogSystem(fmt, args...)
	}
}

func (r PluginRegistry) LogNewSession(sess *Session) {
	for _, p := range loggingPlugins {
		p.LogNewSession(sess)
	}
}

func (r PluginRegistry) LogSessionClosed(sess *Session, err error) {
	for _, p := range loggingPlugins {
		p.LogSessionClosed(sess, err)
	}
}

func (r PluginRegistry) LogClientRequest(sess *Session, msg pgproto.ClientMessage) {
	for _, p := range loggingPlugins {
		p.LogClientRequest(sess, msg)
	}
}

func (r PluginRegistry) LogServerResponse(sess *Session, msg pgproto.ServerMessage) {
	for _, p := range loggingPlugins {
		p.LogServerResponse(sess, msg)
	}
}
