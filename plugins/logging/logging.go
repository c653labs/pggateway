package logging

import (
	"log"

	"github.com/c653labs/pggateway"
	"github.com/c653labs/pgproto"
)

func init() {
	pggateway.RegisterPlugin("logging", &LoggingPlugin{})
}

type LoggingPlugin struct{}

func (l *LoggingPlugin) LogSystem(fmt string, args ...interface{}) {
	log.Printf(fmt, args...)
}

func (l *LoggingPlugin) LogNewSession(sess *pggateway.Session) {
	log.Printf("%s %s %s - new session started", sess.ID, sess.User, sess.Database)
}

func (l *LoggingPlugin) LogSessionClosed(sess *pggateway.Session, err error) {
	log.Printf("%s %s %s - session closed", sess.ID, sess.User, sess.Database)
}

func (l *LoggingPlugin) LogClientRequest(sess *pggateway.Session, msg pgproto.ClientMessage) {
	log.Printf("%s %s %s - client - %s", sess.ID, sess.User, sess.Database, msg)
}

func (l *LoggingPlugin) LogServerResponse(sess *pggateway.Session, msg pgproto.ServerMessage) {
	log.Printf("%s %s %s - server - %s", sess.ID, sess.User, sess.Database, msg)
}
