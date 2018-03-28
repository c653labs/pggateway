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

		go s.handleClient(conn)
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

func (s *Server) handleClient(conn net.Conn) {
	defer conn.Close()
	sess := NewSession(conn)

	err := sess.Negotiate()
	if err != nil {
		log.Println(err)
		return
	}

	// TODO: Have this be real
	//   - sess.startupMsg.User is the username
	if !sess.ValidatePassword([]byte("test")) {
		// TODO: Send error message
		log.Println("password mismatch")
		return
	}

	srv, err := net.Dial("tcp", "127.0.0.1:5432")
	if err != nil {
		log.Println(err)
		return
	}
	defer srv.Close()

	err = sess.Proxy(srv)
	if err != nil {
		log.Println(err)
		return
	}
}
