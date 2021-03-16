package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"runtime/trace"

	"github.com/c653labs/pggateway"
	_ "github.com/c653labs/pggateway/plugins/cloudwatchlogs-logging"
	_ "github.com/c653labs/pggateway/plugins/file-logging"
	_ "github.com/c653labs/pggateway/plugins/iam-authentication"
	_ "github.com/c653labs/pggateway/plugins/passthrough-authentication"
	_ "github.com/c653labs/pggateway/plugins/virtualuser-authentication"
)

var (
	configFilename string
	cpuProfile     string
	traceFile      string
)

func init() {
	flag.StringVar(&configFilename, "config", "pggateway.yaml", "config file to load")
	flag.StringVar(&cpuProfile, "cpuprofile", "", "write cpu profile to file")
	flag.StringVar(&traceFile, "trace", "", "write trace to file")
}

func main() {
	flag.Parse()

	if traceFile != "" {
		f, err := os.Create(traceFile)
		if err != nil {
			log.Fatalf("failed to create trace output file: %v", err)
		}
		defer func() {
			if err := f.Close(); err != nil {
				log.Fatalf("failed to close trace file: %v", err)
			}
		}()

		if err := trace.Start(f); err != nil {
			log.Fatalf("failed to start trace: %v", err)
		}
		defer trace.Stop()
	}

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
