package server

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime/trace"
	"strconv"
	"syscall"

	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"

	"chuanyun.io/esmeralda/controller"
	"chuanyun.io/esmeralda/setting"
	"golang.org/x/sync/errgroup"
)

var exitChan = make(chan int)

var configFilePath = flag.String("config", "/etc/chuanyun/esmeralda.toml", "config file path")

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

func NewServer() Server {
	rootCtx, shutdownFn := context.WithCancel(context.Background())
	childRoutines, childCtx := errgroup.WithContext(rootCtx)

	return &EsmeraldaServerImpl{
		context:       childCtx,
		shutdownFn:    shutdownFn,
		childRoutines: childRoutines,
	}
}

func (this *EsmeraldaServerImpl) Start() {

	go listenToSystemSignals(this)

	setting.Initialize(*configFilePath)

	this.startHttpServer()
}

func (this *EsmeraldaServerImpl) Shutdown(code int, reason string) {
	logrus.Info("Shutdown server started")

	this.shutdownFn()
	this.childRoutines.Wait()

	logrus.WithFields(logrus.Fields{
		"reason": reason,
	}).Info("Shutdown server completed")

	logrus.Exit(code)
}

func (this *EsmeraldaServerImpl) startHttpServer() {

	this.httpServer = NewHttpServer()
	err := this.httpServer.Start(this.context)

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Error("Fail to start server")
		this.Shutdown(1, "Startup failed")
		return
	}
}

type HttpServer struct {
	context context.Context
	httpSrv *http.Server
}

func NewHttpServer() *HttpServer {
	return &HttpServer{}
}

func (this *HttpServer) Start(ctx context.Context) error {
	this.context = ctx

	listenAddr := fmt.Sprintf("%s:%s", "", strconv.FormatInt(setting.Settings.Web.Port, 10))

	router := httprouter.New()
	router.GET(setting.Settings.Web.Prefix+"/", controller.Index)
	router.GET(setting.Settings.Web.Prefix+"/collector/log", controller.Collect)

	router.Handler("GET", setting.Settings.Web.Prefix+"/exporter/metrics", promhttp.Handler())

	this.httpSrv = &http.Server{Addr: listenAddr, Handler: router}

	return this.httpSrv.ListenAndServe()
}

func (this *HttpServer) Shutdown(ctx context.Context) error {
	return this.httpSrv.Shutdown(ctx)
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
