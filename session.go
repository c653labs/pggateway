package pggateway

import (
	"fmt"
	"log"
	"net"

	"github.com/c653labs/pgproto"
	uuid "github.com/satori/go.uuid"
)

type Session struct {
	id       uuid.UUID
	client   net.Conn
	server   net.Conn
	user     []byte
	database []byte
	salt     []byte
	password []byte

	startup *pgproto.StartupMessage
}

func NewSession(client net.Conn, server net.Conn) (*Session, error) {
	var err error
	sess := &Session{
		client: client,
		server: server,
		salt:   generateSalt(),
	}
	sess.id, err = uuid.NewV4()
	if err != nil {
		return nil, err
	}

	return sess, nil
}

func (s *Session) Close() {
	if s.server != nil {
		s.server.Close()
	}
}

func (s *Session) String() string {
	return fmt.Sprintf("Session<ID=%#v, User=%#v, Database=%#v>", s.id.String(), string(s.user), string(s.database))
}

func (s *Session) Handle() error {
	var err error
	s.startup, err = s.parseStartupMessage()
	if err != nil {
		return err
	}

	_, _, err = s.getUserPassword()
	if err != nil {
		return err
	}

	return s.Proxy()
}

func (s *Session) getUserPassword() (*pgproto.AuthenticationRequest, *pgproto.PasswordMessage, error) {
	auth := &pgproto.AuthenticationRequest{
		Method: pgproto.AuthenticationMethodMD5,
		Salt:   s.salt,
	}
	err := s.writeServerMsg(auth)
	if err != nil {
		return nil, nil, err
	}

	msg, err := pgproto.ParseClientMessage(s.client)
	if err != nil {
		return nil, nil, err
	}

	pwdMsg, ok := msg.(*pgproto.PasswordMessage)
	if !ok {
		return nil, nil, fmt.Errorf("expected password message")
	}
	s.password = pwdMsg.Password

	return auth, pwdMsg, nil
}

func (s *Session) parseStartupMessage() (*pgproto.StartupMessage, error) {
	msg, err := pgproto.ParseClientMessage(s.client)
	if err != nil {
		return nil, err
	}

	switch m := msg.(type) {
	case *pgproto.StartupMessage:
		var ok bool
		if s.user, ok = m.Options["user"]; !ok {
			return nil, fmt.Errorf("no username sent with startup message")
		}

		if s.database, ok = m.Options["database"]; !ok {
			return nil, fmt.Errorf("no database name sent with startup message")
		}

		return m, nil
	}
	return nil, fmt.Errorf("unexpected message type")
}

func (s *Session) authenticateWithServer(password []byte) error {
	err := s.writeClientMsg(s.startup)
	if err != nil {
		return err
	}

	msg, err := pgproto.ParseServerMessage(s.server)
	if err != nil {
		return err
	}
	var auth *pgproto.AuthenticationRequest
	var ok bool
	if auth, ok = msg.(*pgproto.AuthenticationRequest); !ok {
		return fmt.Errorf("expected authentication request")
	}

	pwdMsg := &pgproto.PasswordMessage{}
	// Use the salt from the server, not our session salt
	pwdMsg.SetPassword(s.user, password, auth.Salt)
	err = s.writeClientMsg(pwdMsg)
	if err != nil {
		return err
	}

	msg, err = pgproto.ParseServerMessage(s.server)
	if err != nil {
		return err
	}

	auth = nil
	switch m := msg.(type) {
	case *pgproto.AuthenticationRequest:
		auth = m
	case *pgproto.Error:
		// TODO: Write generic cannot connect message?
		return s.writeServerMsg(m)
	default:
		return fmt.Errorf("expected authentication request")
	}

	if auth.Method != pgproto.AuthenticationMethodOK {
		return fmt.Errorf("expected successful authentication request")
	}

	err = s.writeServerMsg(auth)
	log.Printf("%s %s %s - server - %s\r\n", s.id.String(), s.user, s.database, auth)
	return err
}

func (s *Session) Proxy() error {
	// TODO: Use authentication plugin to get password
	err := s.authenticateWithServer([]byte("test"))
	if err != nil {
		return err
	}

	stop := make(chan error)
	go s.proxyClientMessages(stop)
	go s.proxyServerMessages(stop)
	return <-stop
}

func (s *Session) proxyServerMessages(stop chan error) {
	for {
		msg, err := pgproto.ParseServerMessage(s.server)
		if err != nil {
			stop <- err
			break
		}

		s.writeServerMsg(msg)
	}
	stop <- nil
}

func (s *Session) proxyClientMessages(stop chan error) {
	for {
		msg, err := pgproto.ParseClientMessage(s.client)
		if err != nil {
			stop <- err
			break
		}

		s.writeClientMsg(msg)

		if _, ok := msg.(*pgproto.Termination); ok {
			break
		}
	}
	stop <- nil
}

func (s *Session) writeClientMsg(msg pgproto.ClientMessage) error {
	// TODO: Pass through logging plugins
	log.Printf("%s %s %s - client - %s\r\n", s.id.String(), s.user, s.database, msg)
	_, err := msg.WriteTo(s.server)
	return err
}

func (s *Session) writeServerMsg(msg pgproto.ServerMessage) error {
	// TODO: Pass through logging plugins
	log.Printf("%s %s %s - server - %s\r\n", s.id.String(), s.user, s.database, msg)
	_, err := msg.WriteTo(s.client)
	return err
}
