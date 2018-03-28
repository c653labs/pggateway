package pggateway

import (
	"net"
)

type Server struct {
	listener net.Listener
	plugins  PluginRegistry
}

func NewServer() *Server {
	return &Server{
		plugins: NewPluginRegistry(),
	}
}

func (s *Server) acceptConnections() error {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			s.plugins.LogSystem("error accepting client: %s", err)
			return err
		}

		s.plugins.LogSystem("new client session")
		go func() {
			defer conn.Close()
			err := s.handleClient(conn)
			if err != nil {
				s.plugins.LogSystem("error handling client session: %s", err)
			}
		}()
	}
}

func (s *Server) Listen(addr string) error {
	var err error
	s.listener, err = net.Listen("tcp", addr)
	if err != nil {
		s.plugins.LogSystem("error binding to %#v: %s", addr, err)
		return err
	}
	defer s.listener.Close()

	s.plugins.LogSystem("listening for connections: %s", addr)
	return s.acceptConnections()
}

func (s *Server) ListenUnix(addr string) error {
	var err error
	s.listener, err = net.Listen("unix", addr)
	if err != nil {
		s.plugins.LogSystem("error binding to %#v: %s", addr, err)
		return err
	}
	defer s.listener.Close()

	s.plugins.LogSystem("listening for connections: %s", addr)
	return s.acceptConnections()
}

func (s *Server) handleClient(client net.Conn) error {
	server, err := net.Dial("tcp", "127.0.0.1:5432")
	if err != nil {
		s.plugins.LogSystem("error connecting to server %#v: %s", "127.0.0.1:5432", err)
		return err
	}

	sess, err := NewSession(client, server, s.plugins)
	if err != nil {
		s.plugins.LogSystem("error creating new client session: %s", err)
		client.Close()
		return err
	}
	defer sess.Close()

	s.plugins.LogNewSession(sess)
	err = sess.Handle()
	s.plugins.LogSessionClosed(sess, err)
	return err
}