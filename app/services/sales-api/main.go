package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/ardanlabs/conf"
	"github.com/dnmahendra/service/app/services/sales-api/handlers"
	"github.com/pkg/errors"
	"go.uber.org/automaxprocs/maxprocs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

/*
	Need to figure out timeouts for the http service.
*/

// build is the git version of this program. It is set using build flags in the makefile.
var build = "develop"

func main() {
	// Construct the application logger.
	log, err := initLogger("SALES-API")
	if err != nil {
		fmt.Println("Error constructing logger:", err)
		os.Exit(1)
	}
	defer log.Sync()

	// Perform the startup and shutdown sequence
	if err := run(log); err != nil {
		log.Errorw("startup", "ERROR", err)
		os.Exit(1)
	}
}

func run(log *zap.SugaredLogger) error {
	// ========================================================================
	// GOMAXPROCS

	//Set the correct number of threads for the service
	// based on what is available either by the machine or quotas
	if _, err := maxprocs.Set(); err != nil {
		return fmt.Errorf("maxprocs: %w", err)
	}
	log.Infow("startup", "GOMAXPROCS", runtime.GOMAXPROCS(0))

	// ========================================================================
	// Configuration

	cfg := struct {
		conf.Version
		Web struct {
			APIHost         string        `conf:"default:0.0.0.0:6000"`
			DebugHost       string        `conf:"default:0.0.0.0:7000"`
			ReadTimeout     time.Duration `conf:"default:5s"`
			WriteTimeout    time.Duration `conf:"default:5s,noprint"`
			IdleTimeout     time.Duration `conf:"default:120s"`
			ShutdownTimeout time.Duration `conf:"default:5s"`
		}
		DB struct {
			User       string `conf:"default:postgres"`
			Password   string `conf:"default:postgres,noprint"`
			Host       string `conf:"default:db"`
			Name       string `conf:"default:postgres"`
			DisableTLS bool   `conf:"default:false"`
		}
		Auth struct {
			KeyID          string `conf:"default:54bb2165-71e1-41a6-af3e-7da4a0e1e2c1"`
			PrivateKeyFile string `conf:"default:private.pem"`
			Algorithm      string `conf:"default:RS256"`
		}
		Zipkin struct {
			ReporterURI string  `conf:"default:http://zipkin:9411/api/v2/spans"`
			ServiceName string  `conf:"default:sales-api"`
			Probability float64 `conf:"default:0.05"`
		}
	}{
		Version: conf.Version{
			SVN:  build,
			Desc: "copyright information here",
		},
	}

	const prefix = "SALES"
	help, err := conf.ParseOSArgs(prefix, &cfg)
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Println(help)
			return nil
		}

		return fmt.Errorf("parsing config: %w", err)
	}
	// =========================================================================
	// App Starting

	log.Infow("Starting service", "version", build)
	defer log.Infow("Shutdown complete")

	out, err := conf.String(&cfg)
	if err != nil {
		return fmt.Errorf("generating config for output: %w", err)
	}
	log.Infow("Startup", "config", out)

	// ========================================================================
	// Start Debug Service

	log.Infow("startup", "status", "debug router started", "host", cfg.Web.DebugHost)

	// The Debug function returns a mux to listen and serve on for all the debug
	// related endpoints. This include the standard library endpoints.

	// Construct the mux for the debug calls.
	debugMux := handlers.DebugStandardLibraryMux()

	// Start the service listening for debug requests
	// Not concerned with shutting this down with load shedding
	go func() {
		if err := http.ListenAndServe(cfg.Web.DebugHost, debugMux); err != nil {
			log.Errorw("shutdown", "status", "debug router closed", "hist", cfg.Web.DebugHost, "Error", err)
		}
	}()

	// =========================================================================
	// Shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	<-shutdown

	// // ==========================================================================
	// // Initialize authentication support

	// log.Println("main : Started : Initializing authentication support")

	// privatePEM, err := ioutil.ReadFile(cfg.Auth.PrivateKeyFile)
	// if err != nil {
	// 	return errors.Wrap(err, "reading auth private key")
	// }

	// privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privatePEM)
	// if err != nil {
	// 	return errors.Wrap(err, "parsing authe private key")
	// }

	// lookup := func(kid string) (*rsa.PublicKey, error) {
	// 	switch kid {
	// 	case cfg.Auth.KeyID:
	// 		return &privateKey.PublicKey, nil
	// 	}
	// 	return nil, fmt.Errorf("no public key found for the specified kid: %s", kid)
	// }

	// auth, err := auth.New(cfg.Auth.Algorithm, lookup, auth.Keys{cfg.Auth.KeyID: privateKey})
	// if err != nil {
	// 	return errors.Wrap(err, "constructing auth")
	// }

	// // =========================================================================
	// // Start Debug Service
	// //
	// // /debug/pprof - Added to the default mux by importing the net/http/pprof package.
	// // /debug/vars - Added to the default mux by importing the expvar package.
	// //
	// // Not concerned with shutting this down when the application is shutdown.

	// log.Println("main: Initializing debugging support")

	// go func() {
	// 	log.Printf("main: Debug Listening %s", cfg.Web.DebugHost)
	// 	if err := http.ListenAndServe(cfg.Web.DebugHost, http.DefaultServeMux); err != nil {
	// 		log.Printf("main: Debug Listener closed : %v", err)
	// 	}
	// }()

	// // =========================================================================
	// // Start API Service

	// log.Println("main: Initializing API support")

	// // Make a channel to listen for an interrupt or terminate signal from the OS.
	// // Use a buffered channel because the signal package requires it.
	// shutdown := make(chan os.Signal, 1)
	// signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// api := http.Server{
	// 	Addr:         cfg.Web.APIHost,
	// 	Handler:      handlers.API(build, shutdown, log, auth),
	// 	ReadTimeout:  cfg.Web.ReadTimeout,
	// 	WriteTimeout: cfg.Web.WriteTimeout,
	// }

	// // Make a channel to listen for errors coming from the listener. Use a
	// // buffered channel so the goroutine can exit if we don't collect this error.
	// serverErrors := make(chan error, 1)

	// // Start the service listening for requests.
	// go func() {
	// 	log.Printf("main: API listening on %s", api.Addr)
	// 	serverErrors <- api.ListenAndServe()
	// }()

	// // =========================================================================
	// // Shutdown

	// // Blocking main and waiting for shutdown.
	// select {
	// case err := <-serverErrors:
	// 	return errors.Wrap(err, "server error")

	// case sig := <-shutdown:
	// 	log.Printf("main: %v : Start shutdown", sig)

	// 	// Give outstanding requests a deadline for completion.
	// 	ctx, cancel := context.WithTimeout(context.Background(), cfg.Web.ShutdownTimeout)
	// 	defer cancel()

	// 	// Asking listener to shutdown and shed load.
	// 	if err := api.Shutdown(ctx); err != nil {
	// 		api.Close()
	// 		return errors.Wrap(err, "could not stop server gracefully")
	// 	}
	// }

	return nil
}

// Constructs a Sugared Logger that writes to stdout and
// provides human readable timestamps.
func initLogger(service string) (*zap.SugaredLogger, error) {
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"stdout"}
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.DisableStacktrace = true
	config.InitialFields = map[string]interface{}{
		"service": service,
	}

	log, err := config.Build()
	if err != nil {
		return nil, err
	}

	return log.Sugar(), nil
}
