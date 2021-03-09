package pggateway

import (
	"crypto/tls"
	"github.com/c653labs/pgproto"
	"io"
	"net"
)

type Listener struct {
	l        net.Listener
	config   *ListenerConfig
	plugins  *PluginRegistry
	stopping bool
}

func NewListener(config *ListenerConfig) *Listener {
	return &Listener{
		config:   config,
		stopping: false,
	}
}

func (l *Listener) Listen() error {
	l.stopping = false
	var err error
	l.plugins, err = NewPluginRegistry(l.config.Authentication, l.config.Logging)
	if err != nil {
		return err
	}

	l.l, err = net.Listen("tcp", l.config.Bind)
	if err != nil {
		return err
	}

	return nil
}

func (l *Listener) Close() error {
	l.stopping = true
	if l.l != nil {
		l.l.Close()
	}
	return nil
}

func (l *Listener) Handle() error {
	for {
		conn, err := l.l.Accept()
		if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
			continue
		}
		if err != nil {
			if l.stopping {
				return nil
			}
			l.plugins.LogError(nil, "error accepting client: %s", err)
			return err
		}

		go func(conn net.Conn) {
			defer conn.Close()
			err := l.handleClient(conn)
			if err != nil && err != io.EOF {
				l.plugins.LogError(nil, "error handling client session: %s", err)
			}
		}(conn)
	}
}

func (l *Listener) handleClient(client net.Conn) error {

	var err error
	var startup *pgproto.StartupMessage
	var isSSL bool

	startup, err = pgproto.ParseStartupMessage(client)
	if err != nil {
		return err
	}

	if startup.SSLRequest {
		if !l.config.SSL.Enabled {
			_, err = client.Write([]byte{'N'})
			return err
		}
		client, err = l.upgradeSSLConnection(client)
		if err != nil {
			return err
		}
		isSSL = true
		startup, err = pgproto.ParseStartupMessage(client)
		if err != nil {
			return err
		}
	} else if l.config.SSL.Required {
		// SSL is required but they didn't request it, return an error
		return RetunErrorfAndWritePGMsg(client, "server does not support SSL, but SSL was required")
	}

	var user []byte
	var database []byte
	var ok bool

	if user, ok = startup.Options["user"]; !ok {
		// No username was provided
		return RetunErrorfAndWritePGMsg(client, "user startup option is required")
	}

	if database, ok = startup.Options["database"]; !ok {
		// No database was provided
		return RetunErrorfAndWritePGMsg(client, "database startup option is required")
	}

	sess, err := NewSession(startup, user, database, isSSL, client, nil, l.plugins)
	if err != nil {
		l.plugins.LogError(nil, "error creating new client session: %s", err)
		client.Close()
		return err
	}

	defer sess.Close()

	l.plugins.LogInfo(sess.loggingContext(), "new client session")
	err = sess.Handle()

	if err != nil && err != io.EOF {
		l.plugins.LogError(sess.loggingContext(), "client session end: %s", err)
	} else {
		l.plugins.LogInfo(sess.loggingContext(), "client session end")
	}
	return err
}

func (l *Listener) upgradeSSLConnection(client net.Conn) (net.Conn, error) {
	_, err := client.Write([]byte{'S'})
	if err != nil {
		return nil, err
	}

	cer, err := tls.LoadX509KeyPair(l.config.SSL.Certificate, l.config.SSL.Key)
	if err != nil {
		return nil, err
	}

	// Upgrade the client connection to a TLS connection
	sslClient := tls.Server(client, &tls.Config{
		Certificates: []tls.Certificate{cer},
	})
	err = sslClient.Handshake()

	return sslClient, err
}

func (l *Listener) String() string {
	return l.config.Bind
}
