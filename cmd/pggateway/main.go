package main

import (
	"log"
	"net"

	"github.com/c653labs/pggateway"
)

func handleConnection(conn net.Conn) {
	// TODO: How do we know when the connections should be closed?
	// defer conn.Close()

	sess := pggateway.NewSession(conn)
	err := sess.Negotiate()
	if err != nil {
		log.Println(err)
		return
	}

	// TODO: Have this be real
	//   - sess.startupMsg.User is the username
	if !sess.ValidatePassword([]byte("test")) {
		// TODO: Send error message
		log.Println("password mismatch")
		return
	}

	srv, err := net.Dial("tcp", "127.0.0.1:5432")
	if err != nil {
		log.Println(err)
		return
	}
	// defer srv.Close()

	err = sess.Proxy(srv)
	if err != nil {
		log.Println(err)
		return
	}

	// validate username/password against our records
	//   - what types of auth storage are we going to support?
	//     - IAM
	//     - Local/remote db table
	//     - hba/hard coded?
	// connect to the server and authenticate on their behalf
	//   - how are we going to auth?
	//     - multiple user/password which unlocks a single shared db cred for connecting?
	//     - pggateway managed roles on the target db?
	// after sending password message start the proxy between client and server
	//   - we are not going to try and intercept any more messages
	//   - we are not going to handle any parameter status or ready query messages
}

func main() {
	l, err := net.Listen("tcp", ":5433")
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go handleConnection(conn)
	}
}
