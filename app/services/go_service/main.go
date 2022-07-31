package main

import (
	"context"
	"errors"
	"expvar"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/DocLivesey/go_service/app/services/go_service/handlers"
	"github.com/ardanlabs/conf"
	"go.uber.org/automaxprocs/maxprocs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// TODO section
/*
	TODO:
*/

var build = "develop"

func main() {

	//Construct logger
	log, err := initLogger("GO_SERV")
	if err != nil {
		os.Exit(1)
	}
	defer log.Sync()

	//Startup sequences
	if err := run(log); err != nil {
		log.Errorw("startup", "ERROR", err)
		os.Exit(1)
	}

}

func run(log *zap.SugaredLogger) error {

	// Set max go rutines depending on enviroment
	if _, err := maxprocs.Set(); err != nil {
		return fmt.Errorf("maxprocs: %w", err)
	}
	log.Infow("startup", "GOMAXPROCS", runtime.GOMAXPROCS(0))

	//==============================================================================================
	//Configuration

	cfg := struct {
		conf.Version
		Web struct {
			APIHost         string        `conf:"default:0.0.0.0:3000"`
			DebugHost       string        `conf:"default:0.0.0.0:4000"`
			ReadTimeout     time.Duration `conf:"default:5s"`
			WriteTimeout    time.Duration `conf:"default:10s"`
			IdleTimeout     time.Duration `conf:"default:120s"`
			ShutdownTimeout time.Duration `conf:"default:20s"`
		}
	}{
		Version: conf.Version{
			SVN:  build,
			Desc: "copyright",
		},
	}

	const prefix = "SERVICE"
	help, err := conf.ParseOSArgs(prefix, &cfg)

	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Println(help)
			return nil
		}

		return fmt.Errorf("parsing configuration: %w", err)
	}

	log.Infow("starting service", "version", build)
	defer log.Infow("shutdown complete")

	out, err := conf.String(&cfg)
	if err != nil {
		return fmt.Errorf("generating config output : %w", err)
	}
	log.Infow("startup", "config", out)

	expvar.NewString("build").Set(build)

	//==============================================================================================
	//Debug service

	log.Infow("startup", "status", "degug router started", "host", cfg.Web.DebugHost)

	//Debug function returns a mux to listen and serve for all debug related endpoints. This include
	//the standart library endpoints.

	//Construction of the mux for debuging calls
	debugMux := handlers.DebugStandartLibraryMux()

	go func() {
		if err := http.ListenAndServe(cfg.Web.DebugHost, debugMux); err != nil {
			log.Errorw("shutdown", "degug router closed", "host", cfg.Web.DebugHost, "ERROR", err)
		}
	}()

	//=============================================================================================
	// Start API service

	log.Infow("startup", "status", "initializing API service")

	// Make a channel to listen to interupt (Ctrl+C) or terminate signal from os.
	// signal package requires buffered channel
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	api := http.Server{
		Addr:         cfg.Web.APIHost,
		Handler:      nil,
		ReadTimeout:  cfg.Web.ReadTimeout,
		WriteTimeout: cfg.Web.WriteTimeout,
		IdleTimeout:  cfg.Web.IdleTimeout,
		ErrorLog:     zap.NewStdLog(log.Desugar()),
	}

	// Make a channel to listen for errors coming from the listner. Use a buffered channel so the
	// gorutine can exit if we don't collect this error
	serverError := make(chan error, 1)

	// Start the service for listening for api requests.
	go func() {
		log.Infow("startup", "status", "api service started", "host", cfg.Web.APIHost)
		serverError <- api.ListenAndServe()
	}()

	//==============================================================================================
	// Shudown procedure

	// Blocking main and waiting for shutdown
	select {
	case err := <-serverError:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		log.Infow("shutdown", "status", "shutdown started", "signal", sig)
		defer log.Infow("shutdown", "status", "shutdown complete", "signal", sig)

		// Give deadline for dangling responses from our service
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Web.ShutdownTimeout)
		defer cancel()

		// Asking listner to shutdown and shed load
		if err := api.Shutdown(ctx); err != nil {
			api.Close()
			return fmt.Errorf("graceful stop failed : %w", err)
		}
	}
	return nil
}

func initLogger(service string) (*zap.SugaredLogger, error) {

	//Construct application logger
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"stdout"}
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.DisableStacktrace = true
	config.InitialFields = map[string]interface{}{
		"service": service,
	}

	log, err := config.Build()
	if err != nil {
		fmt.Println("Error constructing logger", err)
		return nil, err
	}

	return log.Sugar(), nil
}
