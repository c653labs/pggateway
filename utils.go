package pggateway

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/c653labs/pgproto"
	"github.com/xdg/scram"
	"io"
	"strconv"
	"strings"
)

func generateSalt() []byte {
	salt := make([]byte, 4)
	binary.Read(rand.Reader, binary.BigEndian, &salt[0])
	binary.Read(rand.Reader, binary.BigEndian, &salt[1])
	binary.Read(rand.Reader, binary.BigEndian, &salt[2])
	binary.Read(rand.Reader, binary.BigEndian, &salt[3])
	return salt
}

func FillStruct(data interface{}, result interface{}) error {
	c, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(c, result)
}

func IsDatabaseAllowed(databases []string, database []byte) bool {
	if len(databases) == 0 {
		return true
	}
	for _, match := range databases {
		if match == string(database) {
			return true
		}
	}
	return false
}

func RetunErrorfAndWritePGMsg(out io.Writer, format string, a ...interface{}) error {
	msgString := fmt.Sprintf(format, a...)

	errMsg := &pgproto.Error{
		Severity: []byte("Fatal"),
		Message:  []byte(msgString),
	}
	_, _ = pgproto.WriteMessage(errMsg, out)

	return errors.New(msgString)
}

// CheckMD5UserPassword
func CheckMD5UserPassword(md5UserPassword, salt, md5SumWithSalt []byte) bool {

	digest := md5.New()
	digest.Write(md5UserPassword)
	digest.Write(salt)
	hash := digest.Sum(nil)

	encodedHash := make([]byte, hex.EncodedLen(len(hash)))
	hex.Encode(encodedHash, hash)
	return bytes.Equal(encodedHash, md5SumWithSalt)
}

func GetSCRAMStoredCredentials(scramrolpassword string) (creds scram.StoredCredentials, err error) {
	// strMec, strIter, strSalt, strStorKey, strSrvKey
	s := strings.Split(strings.ReplaceAll(scramrolpassword, "$", ":"), ":")
	if len(s) != 5 {
		return creds, fmt.Errorf("bad rolpassword format")
	}
	i, err := strconv.Atoi(s[1])
	if err != nil {
		return creds, fmt.Errorf("bad iter")
	}
	salt, err := base64.StdEncoding.DecodeString(s[2])
	if err != nil {
		return creds, fmt.Errorf("bad salt")
	}
	storKey, err := base64.StdEncoding.DecodeString(s[3])
	if err != nil {
		return creds, fmt.Errorf("bad storKey")
	}
	servKey, err := base64.StdEncoding.DecodeString(s[4])
	if err != nil {
		return creds, fmt.Errorf("bad servKey")
	}

	return scram.StoredCredentials{
		KeyFactors: scram.KeyFactors{
			Salt:  string(salt),
			Iters: i,
		},
		StoredKey: storKey,
		ServerKey: servKey,
	}, nil
}
