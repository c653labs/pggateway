package pggateway

import (
	"crypto/rand"
	"encoding/binary"
)

func generateSalt() []byte {
	salt := make([]byte, 4)
	binary.Read(rand.Reader, binary.BigEndian, &salt[0])
	binary.Read(rand.Reader, binary.BigEndian, &salt[1])
	binary.Read(rand.Reader, binary.BigEndian, &salt[2])
	binary.Read(rand.Reader, binary.BigEndian, &salt[3])
	return salt
}
