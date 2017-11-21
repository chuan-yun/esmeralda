package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/trace"
	"syscall"

	"chuanyun.io/esmeralda/util"

	"golang.org/x/sync/errgroup"

	"github.com/fsnotify/fsnotify"
	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	commit     = "2000.01.01.release"
	buildstamp = "2000-01-01T00:00:00+0800"
)

var (
	configFilePath = flag.String("config", "/etc/chuanyun/esmeralda.toml", "config file path")
	isShowVersion  = flag.Bool("version", false, "output version information and exit")
	isShowHelp     = flag.Bool("help", false, "output help information and exit")
	pprof          = flag.Bool("pprof", false, "Turn on pprof profiling")
	pprofPort      = flag.Int("pprof.port", 11011, "Define custom port for profiling")

	exitChan = make(chan int)
)

type EsmeraldaServer interface {
	Start()
	Shutdown(code int, reason string)
}

type EsmeraldaServerImpl struct {
	context       context.Context
	shutdownFn    context.CancelFunc
	childRoutines *errgroup.Group
}

func NewEsmeraldaServer() EsmeraldaServer {
	rootCtx, shutdownFn := context.WithCancel(context.Background())
	childRoutines, childCtx := errgroup.WithContext(rootCtx)

	return &EsmeraldaServerImpl{
		context:       childCtx,
		shutdownFn:    shutdownFn,
		childRoutines: childRoutines,
	}
}

func (this *EsmeraldaServerImpl) Start() {
	Config(*configFilePath)
}

func (this *EsmeraldaServerImpl) Shutdown(code int, reason string) {

	// g.log.Info("Shutdown started", "code", code, "reason", reason)

	// err := g.httpServer.Shutdown(g.context)
	// if err != nil {
	// 	g.log.Error("Failed to shutdown server", "error", err)
	// }

	// g.shutdownFn()
	// err = g.childRoutines.Wait()

	// g.log.Info("Shutdown completed", "reason", err)
	// log.Close()
	// os.Exit(code)

}

func printVersionInfo() {
	fmt.Println("esmeralda")
	fmt.Println("commit: " + commit + ", build: " + buildstamp)
	fmt.Println("Copyright (c) 2017, chuanyun.io. All rights reserved.")
}

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.Parse()

	if *isShowVersion {
		printVersionInfo()
		os.Exit(0)
	}

	if *isShowHelp {
		flag.Usage()
		os.Exit(0)
	}

	if *pprof {
		runtime.SetBlockProfileRate(1)
		go func() {
			http.ListenAndServe(fmt.Sprintf("localhost:%d", *pprofPort), nil)
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

	var collectorCmd = &cobra.Command{
		Use:   "collector",
		Short: "Hugo is a very fast static site generator",
		Long: `A Fast and Flexible Static Site Generator built with
					  love by spf13 and friends in Go.
					  Complete documentation is available at http://hugo.spf13.com`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(util.Message(cmd.BashCompletionFunction))
		},
	}

	var transferCmd = &cobra.Command{
		Use:   "transfer",
		Short: "Hugo is a very fast static site generator",
		Long: `A Fast and Flexible Static Site Generator built with
					  love by spf13 and friends in Go.
					  Complete documentation is available at http://hugo.spf13.com`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(util.Message(cmd.BashCompletionFunction))
		},
	}

	var exporterCmd = &cobra.Command{
		Use:   "exporter",
		Short: "Hugo is a very fast static site generator",
		Long: `A Fast and Flexible Static Site Generator built with
					  love by spf13 and friends in Go.
					  Complete documentation is available at http://hugo.spf13.com`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(util.Message(cmd.BashCompletionFunction))
		},
	}

	var apiCmd = &cobra.Command{
		Use:   "api",
		Short: "Hugo is a very fast static site generator",
		Long: `A Fast and Flexible Static Site Generator built with
					  love by spf13 and friends in Go.
					  Complete documentation is available at http://hugo.spf13.com`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(util.Message(cmd.BashCompletionFunction))

			router := httprouter.New()
			router.GET("/", Index)
			router.GET("/hello/:name", Hello)

			panic(http.ListenAndServe(":8080", router))
		},
	}

	collectorCmd.Execute()
	transferCmd.Execute()
	exporterCmd.Execute()
	apiCmd.Execute()

	server := NewEsmeraldaServer()
	server.Start()
}

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "Welcome!\n")
}

func Hello(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprintf(w, "hello, %s!\n", ps.ByName("name"))
}

func exporter() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`
		<html>
			<head><title>Chuanyun Quasimodo Exporter</title></head>
			<body>
				<h1>Chuanyun Quasimodo Exporter</h1>
				<p><a href="/metrics">Metrics</a></p>
			</body>
		</html>`))
	})
	http.Handle("/metrics", promhttp.Handler())
	// logrus.Fatal(http.ListenAndServe(":"+config.Config.Prometheus.Port, nil))
}

func listenToSystemSignals(server EsmeraldaServer) {
	signalChan := make(chan os.Signal, 1)
	ignoreChan := make(chan os.Signal, 1)
	code := 0

	signal.Notify(ignoreChan, syscall.SIGHUP)
	signal.Notify(signalChan, os.Interrupt, os.Kill, syscall.SIGTERM)

	select {
	case sig := <-signalChan:
		// Stops trace if profiling has been enabled
		trace.Stop()
		server.Shutdown(0, util.Message(fmt.Sprintf("system signal: %s", sig)))
	case code = <-exitChan:
		server.Shutdown(code, util.Message("startup error"))
	}
}

func log() {

	// content, err := ioutil.ReadFile("esmeralda.log")
	// if err != nil {
	// 	return nil, err
	// }
	// cfg, err := Load(string(content))
	// if err != nil {
	// 	return nil, err
	// }
	// resolveFilepaths(filepath.Dir(filename), cfg)

	filepath.Base("/a/b.c")

	logrus.SetFormatter(&logrus.JSONFormatter{})

	file, err := os.OpenFile("esmeralda.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err == nil {
		logrus.SetOutput(file)
	} else {
		logrus.SetOutput(os.Stdout)
		logrus.Fatal("Failed to log to file, using default stderr")
	}
	defer file.Close()

	logrus.Debug("Hello World!")
}

func Config(in string) {
	in, err := filepath.Abs(filepath.Clean(in))
	if err != nil {
		panic(util.Message(err.Error()))
	}

	viper.SetEnvPrefix("esmeralda")
	viper.AutomaticEnv()
	viper.SetConfigFile(in)
	viper.SetConfigType("toml")

	err = viper.ReadInConfig()
	if err != nil {
		panic(util.Message(err.Error()))
	}
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		logrus.WithFields(logrus.Fields{
			"filename": e.Name,
		}).Info("Config file changed:")
	})
}
