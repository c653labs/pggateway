package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/c653labs/pggateway"
	_ "github.com/c653labs/pggateway/plugins/logging"
)

func main() {
	// f, err := os.Create("pggateway.pprof")
	// if err != nil {
	//	log.Fatal(err)
	// }
	// pprof.StartCPUProfile(f)
	// defer pprof.StopCPUProfile()

	s := pggateway.NewServer()
	defer s.Close()
	go func() {
		log.Println(s.Listen(":5433"))
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	log.Println("stopping server")
	s.Close()
}
