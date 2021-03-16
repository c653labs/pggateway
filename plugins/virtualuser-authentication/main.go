package virtualuser_authentication

// https://www.postgresql.org/docs/current/protocol.html
// https://www.postgresql.org/docs/current/protocol-flow.html
// https://www.postgresql.org/docs/current/sasl-authentication.html
// https://www.postgresql.org/docs/current/protocol-message-formats.html

import (
	"fmt"
	"github.com/c653labs/pggateway"
	"github.com/c653labs/pgproto"
	"github.com/xdg/scram"
	"strings"
)

// VirtualuserAuthentication
type VirtualuserAuthentications struct {
	UserMap map[string]VirtualuserAuthentication
}

type VirtualuserAuthentication struct {
	Name   string                 `json:"name"`
	Target pggateway.TargetConfig `json:"target"`
	Users  map[string]string      `json:"users"`
}

func init() {
	pggateway.RegisterAuthPlugin("virtualuser-authentication", newVirtualUserPlugin)
}

func newVirtualUserPlugin(config interface{}) (plugin pggateway.AuthenticationPlugin, err error) {
	auths := &[]VirtualuserAuthentication{}
	err = pggateway.FillStruct(config, auths)

	usernameMapping := make(map[string]VirtualuserAuthentication)
	for _, auth := range *auths {
		for username := range auth.Users {
			usernameMapping[username] = auth
		}
	}
	plugin = &VirtualuserAuthentications{
		UserMap: usernameMapping,
	}
	//fmt.Printf("RESEULP: %#v\n\n%#v\n\n%v\n", plugin, config, err)
	return
}

func (p *VirtualuserAuthentications) Authenticate(sess *pggateway.Session) (bool, error) {
	vuauth, ok := p.UserMap[string(sess.User)]
	if !ok {
		return false, fmt.Errorf("virtual user %s does not exist", sess.User)
	}

	if !pggateway.IsDatabaseAllowed(vuauth.Target.Databases, sess.Database) {
		return false, sess.WriteToClientEf("IsDatabaseAllowed returns False")
	}
	err := p.AuthenticateClient(sess)
	if err != nil {
		return false, err
	}
	err = sess.DialToS(vuauth.Target.Host, vuauth.Target.Port)
	if err != nil {
		return false, err
	}
	err = sess.AuthOnServer(vuauth.Target.User, vuauth.Target.Password)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (p *VirtualuserAuthentications) AuthenticateClient(sess *pggateway.Session) (err error) {
	// Client authentication
	customUserName := string(sess.User)
	rolpassword := p.GetRolePassword(customUserName)

	if strings.HasPrefix(rolpassword, "SCRAM-SHA-256$") {
		storedCredentials, err := pggateway.GetSCRAMStoredCredentials(rolpassword)
		if err != nil {
			return fmt.Errorf("cant validate stored creds: %s", err)
		}
		credentiallookup := func(s string) (scram.StoredCredentials, error) {
			// TODO: in SCRAM-...-PLUS will need additional check:
			//if s != customUserName {
			//	return scram.StoredCredentials{}, fmt.Errorf("user not found")
			//}
			return storedCredentials, nil

		}
		err = sess.SCRAMSHA256ClientAuth(credentiallookup)

		return err

	} else if strings.HasPrefix(rolpassword, "md5") {
		authReq, passwd, err := sess.GetUserPassword(pgproto.AuthenticationMethodMD5)
		if err != nil {
			return err
		}
		if !pggateway.CheckMD5UserPassword([]byte(rolpassword[3:]), authReq.Salt, passwd.HeaderMessage[3:]) {
			return fmt.Errorf("failed to login user %s, md5 password check failed", customUserName)
		}
	} else {
		_, passwd, err := sess.GetUserPassword(pgproto.AuthenticationMethodPlaintext)
		if err != nil {
			return fmt.Errorf("failed to get password")
		}
		if string(passwd.HeaderMessage) != rolpassword {
			return fmt.Errorf("failed to login user %s, plaintext password check failed", customUserName)
		}
	}
	return nil
}

func (p *VirtualuserAuthentications) GetRolePassword(username string) string {
	return p.UserMap[username].Users[username]
}
