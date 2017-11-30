package server

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/trace"
	"strconv"
	"syscall"

	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"

	"chuanyun.io/esmeralda/collector"
	"chuanyun.io/esmeralda/controller"
	"chuanyun.io/esmeralda/setting"
	"chuanyun.io/esmeralda/util"
	"golang.org/x/sync/errgroup"
)

var configFilePath = flag.String("config", "/etc/chuanyun/esmeralda.toml", "config file path")
var pidFile = flag.String("pidfile", "", "path to pid file")

type Server interface {
	Start()
	Shutdown(code int, reason string)
}

type EsmeraldaServerImpl struct {
	context       context.Context
	shutdownFn    context.CancelFunc
	childRoutines *errgroup.Group
	httpServer    *HTTPServer
}

func (me *EsmeraldaServerImpl) Start() {

	go listenToSystemSignals(me)

	setting.Initialize(*configFilePath)

	setting.InitializeElasticClient()

	me.writePIDFile()

	me.childRoutines.Go(func() error { return collector.Collector.Run(me.context) })

	me.startHTTPServer()
}

func (me *EsmeraldaServerImpl) Shutdown(code int, reason string) {
	logrus.WithFields(logrus.Fields{
		"code":   code,
		"reason": reason,
	}).Info(util.Message("Shutdown server started"))

	err := me.httpServer.Shutdown(me.context)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Error("Failed to shutdown http server")
	}

	me.shutdownFn()
	if err = me.childRoutines.Wait(); err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Info("Shutdown childRoutines")
	}

	logrus.WithFields(logrus.Fields{
		"reason": reason,
		"code":   code,
	}).Info("Shutdown server completed")

	logrus.Exit(code)
}

func NewEsmeraldaServer() Server {
	rootCtx, shutdownFn := context.WithCancel(context.Background())
	childRoutines, childCtx := errgroup.WithContext(rootCtx)

	return &EsmeraldaServerImpl{
		context:       childCtx,
		shutdownFn:    shutdownFn,
		childRoutines: childRoutines,
	}
}

func (me *EsmeraldaServerImpl) writePIDFile() {
	if *pidFile == "" {
		return
	}

	err := os.MkdirAll(filepath.Dir(*pidFile), 0775)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal(util.Message("Failed to verify pid directory"))
	}

	pid := strconv.Itoa(os.Getpid())
	if err := ioutil.WriteFile(*pidFile, []byte(pid), 0644); err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal(util.Message("Failed to write pidfile"))
	}

	logrus.WithFields(logrus.Fields{
		"path": *pidFile,
		"pid":  pid,
	}).Info("Writing PID file")
}

func (me *EsmeraldaServerImpl) startHTTPServer() {

	me.httpServer = NewHTTPServer()
	err := me.httpServer.Start(me.context)

	if err != nil {
		me.Shutdown(1, "Startup http server failed")

		return
	}
}

type HTTPServer struct {
	context context.Context
	httpSrv *http.Server
}

func (me *HTTPServer) Start(ctx context.Context) (err error) {
	me.context = ctx

	listenAddr := fmt.Sprintf("%s:%s", setting.Settings.Web.Address, strconv.FormatInt(setting.Settings.Web.Port, 10))

	router := httprouter.New()
	router.POST(setting.Settings.Web.Prefix+"/collector/log", collector.HTTPCollector)

	router.Handler("GET", setting.Settings.Web.Prefix+"/metrics", promhttp.Handler())

	logrus.WithFields(logrus.Fields{
		"address":  setting.Settings.Web.Address,
		"port":     setting.Settings.Web.Port,
		"schema":   setting.Settings.Web.Schema,
		"exporter": listenAddr + setting.Settings.Web.Prefix + "/metrics",
	}).Info("Startup http server")

	router.NotFound = http.HandlerFunc(controller.NotFoundHandler)

	me.httpSrv = &http.Server{Addr: listenAddr, Handler: router}

	switch setting.Settings.Web.Schema {
	case setting.HTTP:
		err = me.httpSrv.ListenAndServe()
	default:
		err = errors.New("Invalid web scheme")
	}

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Error(util.Message("http server error"))
	}

	return err
}

func (me *HTTPServer) Shutdown(ctx context.Context) error {
	err := me.httpSrv.Shutdown(ctx)

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Error(util.Message("Fail to shutdown http server"))
	}

	return err
}

func NewHTTPServer() *HTTPServer {
	return &HTTPServer{}
}

func listenToSystemSignals(server Server) {
	signalChan := make(chan os.Signal, 1)
	ignoreChan := make(chan os.Signal, 1)

	signal.Notify(ignoreChan, syscall.SIGHUP)
	signal.Notify(signalChan, os.Interrupt, os.Kill, syscall.SIGTERM)

	select {
	case sig := <-signalChan:
		// Stops trace if profiling has been enabled
		trace.Stop()
		server.Shutdown(0, fmt.Sprintf("system signal: %s", sig))
	}
}
