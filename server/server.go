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

	"chuanyun.io/esmeralda/controller"
	"chuanyun.io/esmeralda/setting"
	"chuanyun.io/esmeralda/util"
	"golang.org/x/sync/errgroup"
)

var exitChan = make(chan int)

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
	httpServer    *HttpServer
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

func (me *EsmeraldaServerImpl) Start() {

	go listenToSystemSignals(me)

	setting.Initialize(*configFilePath)

	me.writePIDFile()
	me.startHttpServer()
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
		}).Error(util.Message("Failed to shutdown http server"))
	}

	me.shutdownFn()
	err = me.childRoutines.Wait()

	logrus.WithFields(logrus.Fields{
		"reason": reason,
		"error":  err,
	}).Info("Shutdown server completed")

	// logrus.Exit(code) will call os.Exit(code)
	logrus.Exit(code)
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

func (me *EsmeraldaServerImpl) startHttpServer() {

	me.httpServer = NewHttpServer()
	err := me.httpServer.Start(me.context)

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Error(util.Message("Fail to start http server"))
		me.Shutdown(1, "Startup http server failed")

		return
	}

	logrus.Info(util.Message("Startup http server success"))
}

type HttpServer struct {
	context context.Context
	httpSrv *http.Server
}

func NewHttpServer() *HttpServer {
	return &HttpServer{}
}

func (me *HttpServer) Start(ctx context.Context) error {
	me.context = ctx

	listenAddr := fmt.Sprintf("%s:%s", "", strconv.FormatInt(setting.Settings.Web.Port, 10))

	router := httprouter.New()
	router.GET(setting.Settings.Web.Prefix+"/", controller.Index)
	router.GET(setting.Settings.Web.Prefix+"/collector/log", controller.Collect)

	router.Handler("GET", setting.Settings.Web.Prefix+"/exporter/metrics", promhttp.Handler())

	me.httpSrv = &http.Server{Addr: listenAddr, Handler: router}

	switch setting.Settings.Web.Schema {
	case setting.HTTP:
		return me.httpSrv.ListenAndServe()
	default:
		return errors.New("Invalid Protocol")
	}
}

func (me *HttpServer) Shutdown(ctx context.Context) error {
	return me.httpSrv.Shutdown(ctx)
}

func listenToSystemSignals(server Server) {
	signalChan := make(chan os.Signal, 1)
	ignoreChan := make(chan os.Signal, 1)
	code := 0

	signal.Notify(ignoreChan, syscall.SIGHUP)
	signal.Notify(signalChan, os.Interrupt, os.Kill, syscall.SIGTERM)

	select {
	case sig := <-signalChan:
		// Stops trace if profiling has been enabled
		trace.Stop()
		server.Shutdown(0, fmt.Sprintf("system signal: %s", sig))
	case code = <-exitChan:
		server.Shutdown(code, "startup error")
	}
}
