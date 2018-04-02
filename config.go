package pggateway

import (
	"github.com/go-yaml/yaml"
)

type Config struct {
	Procs     int                          `yaml:"procs,omitempty"`
	Logging   map[string]map[string]string `yaml:"logging,omitempty"`
	Listeners map[string]*ListenerConfig   `yaml:"listeners,omitempty"`
}

type TargetConfig struct {
	Host    string `yaml:"host,omitempty"`
	Port    int    `yaml:"port,omitempty"`
	SSLMode string `yaml:"sslmode,omitempty"`
}

type SSLConfig struct {
	Enabled     bool   `yaml:"enabled,omitempty"`
	Required    bool   `yaml:"required,omitempty"`
	Certificate string `yaml:"certificate,omitempty"`
	Key         string `yaml:"key,omitempty"`
}

type ListenerConfig struct {
	Bind           string                       `yaml:"bind,omitempty"`
	SSL            SSLConfig                    `yaml:"ssl,omitempty"`
	Target         TargetConfig                 `yaml:"target,omitempty"`
	Authentication map[string]map[string]string `yaml:"authentication,omitempty"`
	Logging        map[string]map[string]string `yaml:"logging,omitempty"`
	Databases      map[string]map[string]string `yaml:"databases,omitempty"`
}

func NewConfig() *Config {
	return &Config{
		Logging:   make(map[string]map[string]string),
		Listeners: make(map[string]*ListenerConfig),
	}
}

func (c *Config) Unmarshal(in []byte) error {
	err := yaml.UnmarshalStrict(in, c)
	if err != nil {
		return err
	}

	return c.resolveListeners()
}

func (c *Config) resolveListeners() error {
	for bind, config := range c.Listeners {
		config.Bind = bind
	}

	return nil
}

func (c *Config) GetListeners() []*Listener {
	listeners := make([]*Listener, 0)
	for _, config := range c.Listeners {
		listeners = append(listeners, NewListener(config))
	}
	return listeners
}
