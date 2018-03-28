package main

import (
	"log"

	"github.com/c653labs/pggateway"
)

func main() {
	s := pggateway.NewServer()
	log.Fatal(s.Listen(":5433"))
}
