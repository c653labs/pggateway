package passthrough

import (
	"github.com/c653labs/pggateway"
)

type Passthrough struct {
	Target pggateway.TargetConfig `json:"target"`
}

func init() {
	pggateway.RegisterAuthPlugin("passthrough", newPassthroughPlugin)
}

func newPassthroughPlugin(config interface{}) (pggateway.AuthenticationPlugin, error) {

	plugin := &Passthrough{}
	err := pggateway.FillStruct(config, plugin)
	return plugin, err
}

func (p *Passthrough) Authenticate(sess *pggateway.Session) (bool, error) {
	if !pggateway.IsDatabaseAllowed(p.Target.Databases, sess.Database) {
		return false, sess.WriteToClientEf("IsDatabaseAllowed returns False")
	}
	err := sess.DialToS(p.Target.Host, p.Target.Port)
	if err != nil {
		return false, err
	}

	return true, sess.WriteToServer(sess.GetStartup())
}
