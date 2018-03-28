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
	msg, err := pgproto.ParseClientMessage(s.conn)
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
		Method: pgproto.AuthenticationMethodMD5,
		Salt:   []byte{'a', 'b', 'c', 'd'},
	}
	_, err := s.conn.Write(s.authReq.Encode())
	if err != nil {
		return err
	}

	msg, err := pgproto.ParseClientMessage(s.conn)
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

	msg, err := pgproto.ParseServerMessage(srv)
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

	msg, err = pgproto.ParseServerMessage(srv)
	if err != nil {
		return err
	}
	ar, ok = msg.(*pgproto.AuthenticationRequest)
	if !ok {
		return fmt.Errorf("expected authentication request")
	}

	if ar.Method != pgproto.AuthenticationMethodOK {
		return fmt.Errorf("expected successful authentication request")
	}

	ar.WriteTo(s.conn)
	log.Printf("Response: %s\r\n", ar)

	// Channel to listen for errors
	stop := make(chan error)
	go func(src io.Reader, dst io.Writer) {
		for {
			msg, err := pgproto.ParseClientMessage(src)
			if err != nil {
				stop <- err
				break
			}

			log.Printf("Request: %s\r\n", msg)
			msg.WriteTo(dst)

			if _, ok := msg.(*pgproto.Termination); ok {
				break
			}
		}
		stop <- nil
	}(s.conn, srv)

	go func(src io.Reader, dst io.Writer) {
		for {
			msg, err := pgproto.ParseServerMessage(src)
			if err != nil {
				stop <- err
				break
			}

			log.Printf("Response: %s\r\n", msg)
			msg.WriteTo(dst)
		}
		stop <- nil
	}(srv, s.conn)

	return <-stop
}
