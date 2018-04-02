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

func newPassthroughPlugin(config map[string]string) (pggateway.AuthenticationPlugin, error) {
	return &Passthrough{}, nil
}

func (p *Passthrough) Authenticate(sess *pggateway.Session, startup *pgproto.StartupMessage) error {
	return sess.WriteToServer(startup)
}
