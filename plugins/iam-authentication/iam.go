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

//goland:noinspection GoNameStartsWithPackageName
type IAMAuth struct {
	RoleArn    string `json:"role"`
	DbUser     string `json:"db"`
	DbPassword string `json:"password"`
	DbSSL      bool   `json:"ssl"`
}

func init() {
	pggateway.RegisterAuthPlugin("iam", newIAMPlugin)
}

func newIAMPlugin(config interface{}) (pggateway.AuthenticationPlugin, error) {
	plugin := &IAMAuth{}
	err := pggateway.FillStruct(config, plugin)
	return plugin, err
}

func (p *IAMAuth) Authenticate(sess *pggateway.Session) (bool, error) {
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
			AccessKeyID:     string(sess.User),
			SecretAccessKey: string(passwd.HeaderMessage),
		}),
	}))
	client := iam.New(awsSess)
	_, err = client.GetUser(nil)
	if err != nil {
		return false, err
	}

	startupReq := &pgproto.StartupMessage{
		SSLRequest: p.DbSSL,
		Options: map[string][]byte{
			"user": []byte(p.DbUser),
		},
	}
	for k, v := range sess.GetStartup().Options {
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
		passwdReq.HeaderMessage = []byte(p.DbPassword)
	case pgproto.AuthenticationMethodMD5:
		passwdReq.SetPassword([]byte(p.DbUser), []byte(p.DbPassword), authResp.Salt)
	default:
		return false, fmt.Errorf("unexpected password request method from server")
	}

	return true, sess.WriteToServer(passwdReq)
}
