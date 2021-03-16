package pggateway

import (
	"crypto/tls"
	"fmt"
	"github.com/xdg/scram"
	"io"
	"net"
	"strconv"
	"sync"

	"github.com/c653labs/pgproto"
	uuid "github.com/satori/go.uuid"
)

type Session struct {
	ID       string
	User     []byte
	Database []byte

	IsSSL    bool
	client   net.Conn
	target   net.Conn
	salt     []byte
	password []byte

	startup *pgproto.StartupMessage

	stopped bool

	plugins *PluginRegistry
}

func NewSession(startup *pgproto.StartupMessage, user []byte, database []byte, isSSL bool, client net.Conn, target net.Conn, plugins *PluginRegistry) (*Session, error) {
	var err error
	id, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	return &Session{
		ID:       id.String(),
		User:     user,
		Database: database,
		IsSSL:    isSSL,
		client:   client,
		target:   target,
		salt:     generateSalt(),
		startup:  startup,
		plugins:  plugins,
		stopped:  false,
	}, nil
}

func (s *Session) Close() {
	if s.target != nil {
		s.target.Close()
	}
}

func (s *Session) String() string {
	return fmt.Sprintf("Session<ID=%#v, User=%#v, Database=%#v>", s.ID, string(s.User), string(s.Database))
}

func (s *Session) Handle() error {
	success, err := s.plugins.Authenticate(s)
	if err != nil {
		return err
	}

	if !success {
		_ = s.WriteToClientEf("failed to authenticate")
		return nil
	}

	return s.proxy()
}

func (s *Session) GetUserPassword(method pgproto.AuthenticationMethod) (*pgproto.AuthenticationRequest, *pgproto.PasswordMessage, error) {
	auth := &pgproto.AuthenticationRequest{
		Method: method,
		Salt:   s.salt,
	}
	err := s.WriteToClient(auth)
	if err != nil {
		return nil, nil, err
	}

	msg, err := s.ParseClientRequest()
	if err != nil {
		return nil, nil, err
	}

	pwdMsg, ok := msg.(*pgproto.PasswordMessage)
	if !ok {
		return nil, nil, fmt.Errorf("expected password message")
	}
	s.password = pwdMsg.HeaderMessage

	return auth, pwdMsg, nil
}

func (s *Session) parseStartupMessage() (*pgproto.StartupMessage, error) {
	msg, err := s.ParseClientRequest()
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

func (s *Session) proxy() error {
	m := &sync.Mutex{}
	stop := sync.NewCond(m)
	errs := make([]error, 0)

	go s.proxyClientMessages(stop, errs)
	go s.proxyServerMessages(stop, errs)

	// Disable message interception
	// go func() {
	//	_, err := io.Copy(s.client, s.target)
	//	errs = append(errs, err)
	//	stop.Broadcast()
	// }()

	// go func() {
	//	_, err := io.Copy(s.target, s.client)
	//	errs = append(errs, err)
	//	stop.Broadcast()
	// }()

	stop.L.Lock()
	stop.Wait()
	stop.L.Unlock()
	s.stopped = true

	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

func (s *Session) proxyServerMessages(stop *sync.Cond, errs []error) {
	var buf []pgproto.Message
	for !s.stopped {
		msg, err := s.ParseServerResponse()
		if err != nil {
			errs = append(errs, err)
			stop.Broadcast()
			break
		}
		buf = append(buf, msg)

		flush := false
		switch m := msg.(type) {
		case *pgproto.ReadyForQuery:
			flush = true
		case *pgproto.AuthenticationRequest:
			flush = m.Method != pgproto.AuthenticationMethodOK
		}
		if flush || len(buf) > 15 {
			pgproto.WriteMessages(buf, s.client)
			buf = nil
		}
	}
	if len(buf) > 0 {
		pgproto.WriteMessages(buf, s.client)
	}
}

func (s *Session) proxyClientMessages(stop *sync.Cond, errs []error) {
	for !s.stopped {
		msg, err := s.ParseClientRequest()
		if err != nil {
			errs = append(errs, err)
			stop.Broadcast()
			break
		}

		s.WriteToServer(msg)

		if _, ok := msg.(*pgproto.Termination); ok {
			break
		}
	}
}

func (s *Session) WriteToServer(msg pgproto.ClientMessage) error {
	_, err := pgproto.WriteMessage(msg, s.target)
	return err
}

func (s *Session) WriteToClient(msg pgproto.ServerMessage) error {
	_, err := pgproto.WriteMessage(msg, s.client)
	return err
}

func (s *Session) ParseClientRequest() (pgproto.ClientMessage, error) {
	msg, err := pgproto.ParseClientMessage(s.client)
	if err == io.EOF {
		return msg, io.EOF
	}

	if err != nil {
		if !s.stopped {
			s.plugins.LogError(s.loggingContextWithMessage(msg), "error parsing client request: %s", err)
		}
	} else {
		s.plugins.LogDebug(s.loggingContextWithMessage(msg), "client request")
	}
	return msg, err
}

func (s *Session) ParseServerResponse() (pgproto.ServerMessage, error) {
	msg, err := pgproto.ParseServerMessage(s.target)
	if err == io.EOF {
		return msg, io.EOF
	}

	if err != nil {
		if !s.stopped {
			s.plugins.LogError(s.loggingContextWithMessage(msg), "error parsing server response: %s", err)
		}
	} else {
		s.plugins.LogDebug(s.loggingContextWithMessage(msg), "server response")
	}
	return msg, err
}

func (s *Session) loggingContext() LoggingContext {
	var cRA, tRA string
	if s.target != nil {
		tRA = s.target.RemoteAddr().String()
	}
	if s.client != nil {
		cRA = s.client.RemoteAddr().String()
	}
	return LoggingContext{
		"session_id": s.ID,
		"user":       string(s.User),
		"database":   string(s.Database),
		"ssl":        s.IsSSL,
		"client":     cRA,
		"target":     tRA,
	}
}

func (s *Session) loggingContextWithMessage(msg pgproto.Message) LoggingContext {
	context := s.loggingContext()
	if msg != nil {
		context["message"] = msg.AsMap()
	}
	return context
}

func (s *Session) WriteToClientEf(format string, a ...interface{}) error {
	return s.WriteToClientEf(format, a...)
}
func (s *Session) ConnectToTarget(addr string) (err error) {

	s.target, err = net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	if s.IsSSL {
		err = s.WriteToServer(&pgproto.SSLRequest{})
		if err != nil {
			return fmt.Errorf("error writing SSLRequest to server: %s", err)
		}
		err = pgproto.ParseSSLResponse(s.target)
		if err != nil {
			return fmt.Errorf("server does not support SSL: %s", err)
		}
		s.target = tls.Client(s.target, &tls.Config{InsecureSkipVerify: true})
	}

	return nil
}

func (s *Session) AuthOnServer(dbUser, dbPassword string) (err error) {

	// Connecting to the postgresql server
	startupReq := &pgproto.StartupMessage{
		SSLRequest: s.IsSSL,
		Options: map[string][]byte{
			"user": []byte(dbUser),
			//"database": []byte(s.Database),
		},
	}
	for k, v := range s.GetStartup().Options {
		if k == "user" {
			continue
		}
		startupReq.Options[k] = v
	}

	err = s.WriteToServer(startupReq)
	if err != nil {
		return err
	}

	srvMsg, err := s.ParseServerResponse()
	if err != nil {
		return err
	}
	authResp, ok := srvMsg.(*pgproto.AuthenticationRequest)
	if !ok {
		return fmt.Errorf("unexpected response type from server request: %s", srvMsg)
	}

	switch authResp.Method {
	case pgproto.AuthenticationMethodOK:
		return s.WriteToClient(authResp)
	case pgproto.AuthenticationMethodPlaintext:
		return s.WriteToServer(&pgproto.PasswordMessage{HeaderMessage: []byte(dbPassword)})
	case pgproto.AuthenticationMethodMD5:
		passwdReq := &pgproto.PasswordMessage{}
		passwdReq.SetPassword([]byte(dbUser), []byte(dbPassword), authResp.Salt)
		return s.WriteToServer(passwdReq)
	case pgproto.AuthenticationMethodSASL:
		err = s.SCRAMSHA256ServerAuth(authResp, dbUser, dbPassword)
		if err != nil {
			return err
		}

		return nil
	default:
		return fmt.Errorf("unexpected password request method from server")
	}
}

func (s *Session) SCRAMSHA256ServerAuth(authResp *pgproto.AuthenticationRequest, dbUser, dbPassword string) (err error) {
	var scramClient *scram.Client
	var scramMechanism, strMsg string
	var rawMsg []byte

	if authResp.SupportedScramSHA256 {
		scramClient, err = scram.SHA256.NewClient(dbUser, dbPassword, "")
		if err != nil {
			return fmt.Errorf("creating %s client %s", pgproto.SASLMechanismScramSHA256, err)
		}
		scramMechanism = pgproto.SASLMechanismScramSHA256

	} else if authResp.SupportedScramSHA256Plus {
		return fmt.Errorf("%s is not supported yet", pgproto.SASLMechanismScramSHA256Plus)
	} else {
		return fmt.Errorf("no found supported sasl mechanisms")
	}

	conv := scramClient.NewConversation()
	strMsg, err = conv.Step("")
	if err != nil {
		return fmt.Errorf("error creating first scram msg %s", err)
	}

	initSASLResponse := &pgproto.PasswordMessage{HeaderMessage: []byte(scramMechanism), BodyMessage: []byte(strMsg)}
	rawMsg, err = s.GetAuthMessageFromServer(initSASLResponse)
	if err != nil {
		return err
	}
	strMsg, err = conv.Step(string(rawMsg))
	if err != nil {
		return fmt.Errorf("second sasl challenge failed: %s", err)
	}

	nextSASLResponse := &pgproto.PasswordMessage{BodyMessage: []byte(strMsg)}
	rawMsg, err = s.GetAuthMessageFromServer(nextSASLResponse)
	if err != nil {
		return err
	}

	strMsg, err = conv.Step(string(rawMsg))

	if err != nil {
		return fmt.Errorf("third sasl challenge failed: %s", err)
	}

	return nil
}

func (s *Session) GetAuthMessageFromServer(message pgproto.ClientMessage) (msg []byte, err error) {

	s.plugins.LogDebug(s.loggingContextWithMessage(message), "gateway request to server")

	err = s.WriteToServer(message)

	if err != nil {
		return
	}

	serverResponse, err := s.ParseServerResponse()

	if err != nil {
		return
	}

	switch response := serverResponse.(type) {

	case *pgproto.AuthenticationRequest:
		return response.Message, nil

	case *pgproto.Error:
		return msg, fmt.Errorf("server responses with error: %s", response.String())

	default:
		return msg, fmt.Errorf("server response is not AuthenticationRequest")
	}
}

func (s *Session) GetPasswordMessageFromClient(auth *pgproto.AuthenticationRequest) ([]byte, error) {
	// it is almost (s *Session) GetUserPassword func
	// but without changing the session context
	s.plugins.LogDebug(s.loggingContextWithMessage(auth), "gateway request to client")

	err := s.WriteToClient(auth)

	if err != nil {
		return nil, fmt.Errorf("write to client error: %s", err)
	}

	msg, err := s.ParseClientRequest()

	if err != nil {
		return nil, fmt.Errorf("parse %T response error: %v", msg, err)
	}

	pwdMsg, ok := msg.(*pgproto.PasswordMessage)

	if !ok {
		return nil, fmt.Errorf("expected PasswordMessage")
	}

	return pwdMsg.BodyMessage, nil
}

func (s *Session) GetStartup() *pgproto.StartupMessage {
	return s.startup
}

func (s *Session) generateSalt() {
	s.salt = generateSalt()
}

func (s *Session) DialToS(host string, port int) error {
	addr := net.JoinHostPort(host, strconv.Itoa(port))
	return s.ConnectToTarget(addr)
}

// Gateway to Client
func (s *Session) SCRAMSHA256ClientAuth(credentiallookup scram.CredentialLookup) error {

	authReq := &pgproto.AuthenticationRequest{
		Method:                   pgproto.AuthenticationMethodSASL,
		SupportedScramSHA256:     true,
		SupportedScramSHA256Plus: false,
	}
	s.plugins.LogDebug(s.loggingContextWithMessage(authReq), "gateway request to client")

	err := s.WriteToClient(authReq)
	if err != nil {
		return fmt.Errorf("write to client error: %s", err)
	}

	clientAuthResp, err := s.ParseClientRequest()
	if err != nil {
		return fmt.Errorf("client does not support sasl auth: %s", err)
	}
	clientAuthMech, ok := clientAuthResp.(*pgproto.PasswordMessage)
	if !ok {
		return fmt.Errorf("client sent not a Password Response: %s", err)
	}
	if len(clientAuthMech.BodyMessage) == 0 {
		return fmt.Errorf("client sent not a SASLInitialResponse: %s", err)
	}
	if string(clientAuthMech.HeaderMessage) != pgproto.SASLMechanismScramSHA256 {
		return fmt.Errorf("client's SASL authentication mechanisms are not supported: %s", err)
	}

	scramServer, err := scram.SHA256.NewServer(credentiallookup)
	if err != nil {
		return fmt.Errorf("failed to start scram conversation server: %s", err)
	}
	conv := scramServer.NewConversation()
	var strMsg string
	strMsg, err = conv.Step(string(clientAuthMech.BodyMessage))
	if err != nil {
		return fmt.Errorf("e=unknown-user: %s", err)
	}
	clientResp, err := s.GetPasswordMessageFromClient(
		&pgproto.AuthenticationRequest{
			Method:  pgproto.AuthenticationMethodSASLContinue,
			Message: []byte(strMsg),
		})
	if err != nil {
		return fmt.Errorf("AuthenticationMethodSASLContinue error: %s", err)
	}
	strMsg, err = conv.Step(string(clientResp))
	if err != nil {
		// strMsg == "e=invalid-proof"
		s.WriteToClientEf("failed to authenticate user %s", string(s.User))
		return fmt.Errorf("auth failed: %s [%s]", strMsg, err)
	}
	err = s.WriteToClient(&pgproto.AuthenticationRequest{
		Method:  pgproto.AuthenticationMethodSASLFinal,
		Message: []byte(strMsg),
	})
	if err != nil {
		return fmt.Errorf("AuthenticationMethodSASLFinal error: %s", err)
	}
	return nil
}
