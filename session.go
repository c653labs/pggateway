package pggateway

import (
	"fmt"
	"io"
	"log"
	"net"

	"github.com/c653labs/pgproto"
)

type Session struct {
	conn       net.Conn
	startupMsg *pgproto.StartupMessage
	authReq    *pgproto.AuthenticationRequest
	pwdMsg     *pgproto.PasswordMessage
}

func NewSession(conn net.Conn) *Session {
	return &Session{conn: conn}
}

func (s *Session) Negotiate() error {
	msg, err := pgproto.ParseMessage(s.conn)
	if err != nil {
		return err
	}

	switch m := msg.(type) {
	case *pgproto.StartupMessage:
		return s.handleStartup(m)
	}
	return fmt.Errorf("unexpected message type")
}

func (s *Session) ValidatePassword(password []byte) bool {
	return s.pwdMsg.PasswordValid(s.startupMsg.Options["user"], password, s.authReq.Salt)
}

func (s *Session) handleStartup(startup *pgproto.StartupMessage) error {
	s.startupMsg = startup

	s.authReq = &pgproto.AuthenticationRequest{
		Method: pgproto.AUTHENTICATION_MD5,
		Salt:   []byte{'a', 'b', 'c', 'd'},
	}
	_, err := s.conn.Write(s.authReq.Encode())
	if err != nil {
		return err
	}

	msg, err := pgproto.ParseMessage(s.conn)
	if err != nil {
		return err
	}

	var ok bool
	s.pwdMsg, ok = msg.(*pgproto.PasswordMessage)
	if !ok {
		return fmt.Errorf("expected password message")
	}

	return nil
}

func (s *Session) Proxy(srv net.Conn) error {
	_, err := s.startupMsg.WriteTo(srv)
	if err != nil {
		return err
	}

	msg, err := pgproto.ParseMessage(srv)
	if err != nil {
		return err
	}
	ar, ok := msg.(*pgproto.AuthenticationRequest)
	if !ok {
		return fmt.Errorf("expected authentication request")
	}

	pwdHash := &pgproto.PasswordMessage{}
	pwdHash.SetPassword(s.startupMsg.Options["user"], []byte("test"), ar.Salt)
	_, err = pwdHash.WriteTo(srv)
	if err != nil {
		return err
	}

	msg, err = pgproto.ParseMessage(srv)
	if err != nil {
		return err
	}
	ar, ok = msg.(*pgproto.AuthenticationRequest)
	if !ok {
		return fmt.Errorf("expected authentication request")
	}

	if ar.Method != pgproto.AUTHENTICATION_OK {
		return fmt.Errorf("expected successful authentication request")
	}

	ar.WriteTo(s.conn)
	// go io.Copy(s.conn, srv)
	// go io.Copy(srv, s.conn)
	go func(src io.Reader, dst io.Writer) {
		for {
			msg, err := pgproto.ParseMessage(src)
			if err != nil {
				log.Printf("%v\r\n", err)
				break
			}

			log.Printf("Request: %s\r\n", msg)
			msg.WriteTo(dst)
		}
	}(s.conn, srv)

	go func(src io.Reader, dst io.Writer) {
		for {
			msg, err := pgproto.ParseMessage(src)
			if err != nil {
				log.Printf("Response: %v\r\n", err)
				break
			}

			log.Printf("Response: %s\r\n", msg)
			msg.WriteTo(dst)
		}
	}(srv, s.conn)

	return nil
}
