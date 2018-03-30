package pggateway

import (
	"io"
	"net"
)

type Server struct {
	listener net.Listener
	plugins  *PluginRegistry
}

func NewServer() (*Server, error) {
	registry, err := NewPluginRegistry()
	return &Server{
		plugins: registry,
	}, err
}

func (s *Server) acceptConnections() error {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			s.plugins.LogError(nil, "error accepting client: %s", err)
			return err
		}

		s.plugins.LogInfo(nil, "new client session")
		go func() {
			defer conn.Close()
			err := s.handleClient(conn)
			if err != nil && err != io.EOF {
				s.plugins.LogError(nil, "error handling client session: %s", err)
			}
		}()
	}
}

func (s *Server) Listen(addr string) error {
	var err error
	s.listener, err = net.Listen("tcp", addr)
	if err != nil {
		s.plugins.LogError(nil, "error binding to %#v: %s", addr, err)
		return err
	}

	s.plugins.LogWarn(nil, "listening for connections: %s", addr)
	return s.acceptConnections()
}

func (s *Server) ListenUnix(addr string) error {
	var err error
	s.listener, err = net.Listen("unix", addr)
	if err != nil {
		s.plugins.LogError(nil, "error binding to %#v: %s", addr, err)
		return err
	}

	s.plugins.LogInfo(nil, "listening for connections: %s", addr)
	return s.acceptConnections()
}

func (s *Server) Close() error {
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

func (s *Server) handleClient(client net.Conn) error {
	server, err := net.Dial("tcp", "127.0.0.1:5432")
	if err != nil {
		s.plugins.LogError(nil, "error connecting to server %#v: %s", "127.0.0.1:5432", err)
		return err
	}

	sess, err := NewSession(client, server, s.plugins)
	if err != nil {
		s.plugins.LogError(nil, "error creating new client session: %s", err)
		client.Close()
		return err
	}
	defer sess.Close()

	s.plugins.LogInfo(sess.loggingContext(), "new client session")
	err = sess.Handle()

	s.plugins.LogInfo(sess.loggingContext(), "%s", err)
	return err
}
