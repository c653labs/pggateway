package main

import (
	"log"

	"github.com/c653labs/pggateway"
	_ "github.com/c653labs/pggateway/plugins/logging"
)

func main() {
	s := pggateway.NewServer()
	log.Fatal(s.Listen(":5433"))
}
