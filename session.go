package pggateway

import (
	"crypto/tls"
	"fmt"
	"net"

	"github.com/c653labs/pgproto"
	uuid "github.com/satori/go.uuid"
)

type Session struct {
	ID       string
	User     []byte
	Database []byte

	client   net.Conn
	server   net.Conn
	salt     []byte
	password []byte

	startup *pgproto.StartupMessage

	plugins PluginRegistry
}

func NewSession(client net.Conn, server net.Conn, plugins PluginRegistry) (*Session, error) {
	var err error
	id, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	return &Session{
		ID:      id.String(),
		client:  client,
		server:  server,
		salt:    generateSalt(),
		plugins: plugins,
	}, nil
}

func (s *Session) Close() {
	if s.server != nil {
		s.server.Close()
	}
}

func (s *Session) String() string {
	return fmt.Sprintf("Session<ID=%#v, User=%#v, Database=%#v>", s.ID, string(s.User), string(s.Database))
}

func (s *Session) Handle() error {
	var err error
	s.startup, err = s.parseStartupMessage()
	if err != nil {
		return err
	}

	if s.startup.SSLRequest {
		return s.setupSSLConnection()
	}

	_, _, err = s.getUserPassword()
	if err != nil {
		return err
	}

	return s.proxy()
}

func (s *Session) setupSSLConnection() error {
	_, err := s.client.Write([]byte{'S'})
	if err != nil {
		return err
	}

	cer, err := tls.LoadX509KeyPair("server.crt", "server.key")
	if err != nil {
		return err
	}

	// Upgrade the client connection to a TLS connection
	s.client = tls.Server(s.client, &tls.Config{
		Certificates: []tls.Certificate{cer},
	})

	return s.Handle()
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

	msg, err := s.parseClientMessage()
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
	msg, err := s.parseClientMessage()
	if err != nil {
		return nil, err
	}

	switch m := msg.(type) {
	case *pgproto.StartupMessage:
		// Only extract options if this isn't an SSL request
		if m.SSLRequest {
			return m, nil
		}

		var ok bool
		if s.User, ok = m.Options["user"]; !ok {
			return nil, fmt.Errorf("no username sent with startup message")
		}

		if s.Database, ok = m.Options["database"]; !ok {
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

	msg, err := s.parseServerMessage()
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
	pwdMsg.SetPassword(s.User, password, auth.Salt)
	err = s.writeClientMsg(pwdMsg)
	if err != nil {
		return err
	}

	msg, err = s.parseServerMessage()
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
	return err
}

func (s *Session) proxy() error {
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
		msg, err := s.parseServerMessage()
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
		msg, err := s.parseClientMessage()
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
	_, err := msg.WriteTo(s.server)
	return err
}

func (s *Session) writeServerMsg(msg pgproto.ServerMessage) error {
	_, err := msg.WriteTo(s.client)
	return err
}

func (s *Session) parseClientMessage() (pgproto.ClientMessage, error) {
	msg, err := pgproto.ParseClientMessage(s.client)
	if err != nil {
		s.plugins.LogSystem("error parsing client message: %s", err)
	} else {
		s.plugins.LogClientRequest(s, msg)
	}
	return msg, err
}

func (s *Session) parseServerMessage() (pgproto.ServerMessage, error) {
	msg, err := pgproto.ParseServerMessage(s.server)
	if err != nil {
		s.plugins.LogSystem("error parsing server message: %#v", err)
	} else {
		s.plugins.LogServerResponse(s, msg)
	}
	return msg, err
}
