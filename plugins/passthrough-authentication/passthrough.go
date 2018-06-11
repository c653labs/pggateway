package passthrough

import (
	"github.com/c653labs/pggateway"
	"github.com/c653labs/pgproto"
)

type Passthrough struct {
}

func init() {
	pggateway.RegisterAuthPlugin("passthrough", newPassthroughPlugin)
}

func newPassthroughPlugin(config pggateway.ConfigMap) (pggateway.AuthenticationPlugin, error) {
	return &Passthrough{}, nil
}

func (p *Passthrough) Authenticate(sess *pggateway.Session, startup *pgproto.StartupMessage) (bool, error) {
	return true, sess.WriteToServer(startup)
}
