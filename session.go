package pggateway

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/c653labs/pgmsg"
)

type Session struct {
	conn       net.Conn
	startupMsg *pgmsg.StartupMessage
	authReq    *pgmsg.AuthenticationRequest
	pwdMsg     *pgmsg.PasswordMessage
}

func NewSession(conn net.Conn) *Session {
	return &Session{conn: conn}
}

func (s *Session) Negotiate() error {
	msg, err := pgmsg.ParseMessage(s.conn)
	if err != nil {
		return err
	}

	switch m := msg.(type) {
	case *pgmsg.StartupMessage:
		return s.handleStartup(m)
	}
	return fmt.Errorf("unexpected message type")
}

func (s *Session) ValidatePassword(password []byte) bool {
	hash := hashPassword([]byte(s.startupMsg.Options["user"]), password, s.authReq.Salt)
	return bytes.Equal(s.pwdMsg.Password, hash)
}

func (s *Session) handleStartup(startup *pgmsg.StartupMessage) error {
	s.startupMsg = startup

	s.authReq = &pgmsg.AuthenticationRequest{
		Method: pgmsg.AUTHENTICATION_MD5,
		Salt:   []byte{'a', 'b', 'c', 'd'},
	}
	_, err := s.conn.Write(s.authReq.Encode())
	if err != nil {
		return err
	}

	msg, err := pgmsg.ParseMessage(s.conn)
	if err != nil {
		return err
	}

	var ok bool
	s.pwdMsg, ok = msg.(*pgmsg.PasswordMessage)
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

	msg, err := pgmsg.ParseMessage(srv)
	if err != nil {
		return err
	}
	ar, ok := msg.(*pgmsg.AuthenticationRequest)
	if !ok {
		return fmt.Errorf("expected authentication request")
	}

	pwdHash := &pgmsg.PasswordMessage{
		Password: hashPassword([]byte(s.startupMsg.Options["user"]), []byte("test"), ar.Salt),
	}
	_, err = pwdHash.WriteTo(srv)
	if err != nil {
		return err
	}

	msg, err = pgmsg.ParseMessage(srv)
	if err != nil {
		return err
	}
	ar, ok = msg.(*pgmsg.AuthenticationRequest)
	if !ok {
		return fmt.Errorf("expected authentication request")
	}

	if ar.Method != pgmsg.AUTHENTICATION_OK {
		return fmt.Errorf("expected successful authentication request")
	}

	ar.WriteTo(s.conn)
	// go io.Copy(s.conn, srv)
	// go io.Copy(srv, s.conn)
	go func(src io.Reader, dst io.Writer) {
		for {
			msg, err := pgmsg.ParseMessage(src)
			if err != nil {
				log.Printf("%#v\r\n", err)
				break
			}

			log.Printf("Request: %#v\r\n", msg)
			msg.WriteTo(dst)
		}
	}(s.conn, srv)

	go func(src io.Reader, dst io.Writer) {
		for {
			msg, err := pgmsg.ParseMessage(src)
			if err != nil {
				log.Printf("Response: %#v\r\n", err)
				break
			}

			log.Printf("Response: %#v\r\n", msg)
			msg.WriteTo(dst)
		}
	}(srv, s.conn)

	return nil
}
