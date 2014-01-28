package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"net/http"
	_ "net/http/pprof"
	_ "github.com/lox/opencoindata/command"
	"github.com/mitchellh/cli"
)

func main() {
	os.Exit(realMain())
}

func realMain() int {
	if os.Getenv("PROFILE") == "1" {
		go func() {
			ticker := time.NewTicker(time.Second * 10)
			for _ = range ticker.C {
				runtime.GC()
				var s runtime.MemStats
				runtime.ReadMemStats(&s)
				fmt.Printf("Alloc: %d Sys: %d Gc: %d GoRoutines: %d\n",
					s.Alloc, s.Sys, s.NumGC, runtime.NumGoroutine())
			}
		}()

		log.Printf("Profiling on http://localhost:6060/debug/pprof")
		go func() {
			log.Println(http.ListenAndServe("localhost:6060", nil))
		}()
	}

	// If there is no explicit number of Go threads to use, then set it
	if os.Getenv("GOMAXPROCS") == "" {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	// log.SetOutput(ioutil.Discard)
	// show the version
	args := os.Args[1:]
	for _, arg := range args {
		if arg == "-v" || arg == "--version" {
			newArgs := make([]string, len(args)+1)
			newArgs[0] = "version"
			copy(newArgs[1:], args)
			args = newArgs
			break
		}
	}

	cli := &cli.CLI{
		Args:     args,
		Commands: Commands,
	}

	exitCode, err := cli.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing CLI: %s\n", err.Error())
		return 1
	}

	return exitCode
}
