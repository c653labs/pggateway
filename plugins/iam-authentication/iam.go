package iam

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/c653labs/pggateway"
	"github.com/c653labs/pgproto"
)

type IAMAuth struct {
	roleArn    string
	dbUser     string
	dbPassword string
	dbSSL      bool
}

func init() {
	pggateway.RegisterAuthPlugin("iam", newIAMPlugin)
}

func newIAMPlugin(config pggateway.ConfigMap) (pggateway.AuthenticationPlugin, error) {
	var ok bool
	auth := &IAMAuth{}

	auth.roleArn, ok = config.String("role")
	if !ok {
		return nil, fmt.Errorf("'role' configuration value is required")
	}

	db, ok := config.Map("db")
	if !ok {
		return nil, fmt.Errorf("'db' configuration value is required")
	}

	auth.dbUser, ok = db.String("user")
	if !ok {
		return nil, fmt.Errorf("'db.user' configuration value is required")
	}
	auth.dbPassword = db.StringDefault("password", "")
	auth.dbSSL = db.BoolDefault("ssl", true)

	return auth, nil
}

func (p *IAMAuth) Authenticate(sess *pggateway.Session, startup *pgproto.StartupMessage) (bool, error) {
	// We are passing through IAM credentials... don't let people do silly things
	if !sess.IsSSL {
		return false, fmt.Errorf("IAM auth requires an SSL session")
	}

	_, passwd, err := sess.GetUserPassword(pgproto.AuthenticationMethodPlaintext)
	if err != nil {
		return false, err
	}

	awsSess := session.Must(session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentialsFromCreds(credentials.Value{
			AccessKeyID:     string(startup.Options["user"]),
			SecretAccessKey: string(passwd.Password),
		}),
	}))
	client := iam.New(awsSess)
	_, err = client.GetUser(nil)
	if err != nil {
		return false, err
	}

	startupReq := &pgproto.StartupMessage{
		SSLRequest: p.dbSSL,
		Options: map[string][]byte{
			"user": []byte(p.dbUser),
		},
	}
	for k, v := range startup.Options {
		if k == "user" {
			continue
		}
		startupReq.Options[k] = v
	}
	err = sess.WriteToServer(startupReq)
	if err != nil {
		return false, err
	}

	srvMsg, err := sess.ParseServerResponse()
	if err != nil {
		return false, err
	}
	authResp, ok := srvMsg.(*pgproto.AuthenticationRequest)
	if !ok {
		return false, fmt.Errorf("unexpected response type from server request")
	}
	if authResp.Method == pgproto.AuthenticationMethodOK {
		return true, sess.WriteToClient(authResp)
	}

	passwdReq := &pgproto.PasswordMessage{}
	switch authResp.Method {
	case pgproto.AuthenticationMethodPlaintext:
		passwdReq.Password = []byte(p.dbPassword)
	case pgproto.AuthenticationMethodMD5:
		passwdReq.SetPassword([]byte(p.dbUser), []byte(p.dbPassword), authResp.Salt)
	default:
		return false, fmt.Errorf("unexpected password request method from server")
	}

	return true, sess.WriteToServer(passwdReq)
}
