package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"runtime/trace"

	_ "net/http/pprof"

	"github.com/chuan-yun/esmeralda/server"
)

var (
	commit     = "N/A"
	buildstamp = "N/A"
)

var (
	isShowVersionInfo = flag.Bool("version", false, "output version information and exit")
	isShowHelpInfo    = flag.Bool("help", false, "output help information and exit")
	profiling         = flag.Bool("pprof", false, "Turn on pprof profiling")
	profilingPort     = flag.Int("pprof.port", 11011, "Define custom port for pprof profiling")
)

func main() {

	flag.Parse()

	if *isShowVersionInfo {
		fmt.Println(filepath.Base(os.Args[0]))
		fmt.Println("commit: " + commit + ", build: " + buildstamp)
		fmt.Println("Copyright (c) 2017, chuanyun.io. All rights reserved.")
		os.Exit(0)
	}

	if *isShowHelpInfo {
		flag.Usage()
		os.Exit(0)
	}

	if *profiling {
		runtime.SetBlockProfileRate(1)
		go func() {
			http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", *profilingPort), nil)
		}()

		f, err := os.Create("trace.out")
		if err != nil {
			panic(err)
		}
		defer f.Close()

		err = trace.Start(f)
		if err != nil {
			panic(err)
		}
		defer trace.Stop()
	}

	server := server.NewEsmeraldaServer()
	server.Start()
}
