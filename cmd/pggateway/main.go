package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"

	"github.com/c653labs/pggateway"
	_ "github.com/c653labs/pggateway/plugins/cloudwatchlogs-logging"
	_ "github.com/c653labs/pggateway/plugins/file-logging"
	_ "github.com/c653labs/pggateway/plugins/passthrough-authentication"
)

var (
	configFilename string
	cpuProfile     string
)

func init() {
	flag.StringVar(&configFilename, "config", "pggateway.yaml", "config file to load")
	flag.StringVar(&cpuProfile, "cpuprofile", "", "write cpu profile to file")
}

func main() {
	flag.Parse()

	if cpuProfile != "" {
		f, err := os.Create(cpuProfile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	c := pggateway.NewConfig()
	cf, err := ioutil.ReadFile(configFilename)
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
		if err != nil {
			log.Fatalf("error starting: %#v", err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig
}
