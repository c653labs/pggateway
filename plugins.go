package pggateway

var authPlugin AuthenticationPlugin
var loggingPlugins = make(map[string]LoggingPlugin)

type Plugin interface{}

type AuthenticationPlugin interface {
	Plugin
	OnStart()
	OnValidate()
	OnFinalize()
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

func (r PluginRegistry) LogInfo(context LoggingContext, msg string, args ...interface{}) {
	for _, p := range loggingPlugins {
		go p.LogInfo(context, msg, args...)
	}
}

func (r PluginRegistry) LogError(context LoggingContext, msg string, args ...interface{}) {
	for _, p := range loggingPlugins {
		go p.LogError(context, msg, args...)
	}
}

func (r PluginRegistry) LogWarn(context LoggingContext, msg string, args ...interface{}) {
	for _, p := range loggingPlugins {
		go p.LogWarn(context, msg, args...)
	}
}

func (r PluginRegistry) LogDebug(context LoggingContext, msg string, args ...interface{}) {
	for _, p := range loggingPlugins {
		go p.LogDebug(context, msg, args...)
	}
}

func (r PluginRegistry) LogFatal(context LoggingContext, msg string, args ...interface{}) {
	for _, p := range loggingPlugins {
		go p.LogFatal(context, msg, args...)
	}
}
