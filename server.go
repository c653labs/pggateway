package pggateway

import "sync"

type Server struct {
	listeners []*Listener
	plugins   *PluginRegistry
	config    *Config
}

func NewServer(c *Config) (*Server, error) {
	registry, err := NewPluginRegistry(nil, c.Logging)
	if err != nil {
		return nil, err
	}
	return &Server{
		listeners: make([]*Listener, 0),
		plugins:   registry,
		config:    c,
	}, nil
}

func (s *Server) Start() error {
	m := &sync.Mutex{}
	c := sync.NewCond(m)
	errs := make([]error, 0)

	s.listeners = s.config.GetListeners()
	for _, l := range s.listeners {
		err := l.Listen()
		if err != nil {
			s.plugins.LogError(nil, "error binding to %s: %s", l, err)
			return err
		}

		s.plugins.LogWarn(nil, "listening for connections: %v", l.String())
		go func(l *Listener) {
			err := l.Handle()
			errs = append(errs, err)
			c.Broadcast()
		}(l)
	}

	c.L.Lock()
	c.Wait()
	c.L.Unlock()
	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

func (s *Server) Close() error {
	s.plugins.LogWarn(nil, "stopping server")
	var err error
	for _, l := range s.listeners {
		e := l.Close()
		if e != nil {
			err = e
		}
	}
	return err
}
