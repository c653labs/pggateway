package pggateway

import (
	"log"
	"net"
)

type Server struct {
	listener net.Listener
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) acceptConnections() error {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return err
		}

		go func() {
			defer conn.Close()
			err := s.handleClient(conn)
			if err != nil {
				log.Println(err)
			}
		}()
	}
}

func (s *Server) Listen(addr string) error {
	var err error
	s.listener, err = net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer s.listener.Close()

	return s.acceptConnections()
}

func (s *Server) ListenUnix(addr string) error {
	var err error
	s.listener, err = net.Listen("unix", addr)
	if err != nil {
		return err
	}
	defer s.listener.Close()

	return s.acceptConnections()
}

func (s *Server) handleClient(client net.Conn) error {
	server, err := net.Dial("tcp", "127.0.0.1:5432")
	if err != nil {
		return err
	}

	sess, err := NewSession(client, server)
	if err != nil {
		client.Close()
		return err
	}
	defer sess.Close()

	return sess.Handle()
}
