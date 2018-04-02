package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"runtime"

	"github.com/c653labs/pggateway"
	_ "github.com/c653labs/pggateway/plugins/logging"
	_ "github.com/c653labs/pggateway/plugins/passthrough-auth"
)

func main() {
	c := pggateway.NewConfig()
	cf, err := ioutil.ReadFile("example.yaml")
	if err != nil {
		log.Fatal(err)
	}

	err = c.Unmarshal(cf)
	if err != nil {
		log.Fatal(err)
	}

	// Set max number of CPUs to use for goroutines
	if c.Procs > 0 {
		runtime.GOMAXPROCS(c.Procs)
	}

	s, err := pggateway.NewServer(c)
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()
	go func() {
		err = s.Start()
		log.Printf("error starting: %#v", err)
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig

	log.Println("stopping server")
	s.Close()
}
