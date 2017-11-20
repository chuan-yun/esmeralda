package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"runtime/trace"

	"golang.org/x/sync/errgroup"

	"github.com/fsnotify/fsnotify"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	commit     = "2000.01.01.release"
	buildstamp = "2000-01-01T00:00:00+0800"
)

var (
	configFile    = flag.String("config", "/etc/chuanyun/esmeralda.toml", "config file path")
	isShowVersion = flag.Bool("version", false, "output version information and exit")
	isShowHelp    = flag.Bool("help", false, "output help information and exit")
	pprof         = flag.Bool("pprof", false, "Turn on pprof profiling")
	pprofPort     = flag.Int("pprof.port", 11011, "Define custom port for profiling")
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
}

func (this *EsmeraldaServerImpl) Shutdown(code int, reason string) {
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

	fmt.Println(*config)

	dir := filepath.Dir(*config)
	fmt.Print("Dir=")
	fmt.Println(dir)

	dir, err := filepath.Abs(filepath.Clean(filepath.Dir(*config)))
	if err != nil {
		logrus.Fatal(err)
	}
	fmt.Print("Abs=")
	fmt.Println(dir)

	dir, err = os.Getwd()
	fmt.Print("Wd=")
	fmt.Println(dir)
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

func Config() {
	viper.SetEnvPrefix("esmeralda")
	viper.AutomaticEnv()

	viper.SetConfigType("toml")
	viper.SetConfigName("esmeralda")
	viper.AddConfigPath("/etc/chuanyun/")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Panic("error occurred during config initialization")
	}
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		logrus.WithFields(logrus.Fields{
			"filename": e.Name,
		}).Info("Config file changed:")
	})

	logrus.WithFields(logrus.Fields{
		"settings": viper.AllSettings(),
	}).Info("all user settings")
}
