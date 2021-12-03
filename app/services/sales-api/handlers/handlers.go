// Package handlers contains the full set of handler functions and routes
// supported by the web api.
package handlers

import (
	"expvar"
	"log"
	"net/http"
	"net/http/pprof"
	"os"

	"github.com/dnmahendra/service/business/auth"
	"github.com/dnmahendra/service/business/mid"
	"github.com/dnmahendra/service/foundation/web"
)

// DebugStandardLibraryMux registers all the debug routes from the standard library
// into a new mux bypassing the use of the DefaultServerMux. Using the DefaultServerMux
// would be a security risk since a dependency could inject a handler into our service
// without us knowing it.
func DebugStandardLibraryMux() *http.ServeMux {
	mux := http.NewServeMux()

	// Register all the standard library debug endpoints
	mux.HandleFunc("/debug/pprof", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	mux.Handle("/debug/vars", expvar.Handler())

	return mux
}

// API constructs an http.Handler with all application routes defined.
func API(build string, shutdown chan os.Signal, log *log.Logger, a *auth.Auth) *web.App {
	app := web.NewApp(shutdown, mid.Logger(log), mid.Errors(log), mid.Metrics(), mid.Panics(log))

	check := check{
		log: log,
	}
	app.Handle(
		http.MethodGet,
		"/readiness",
		check.readiness,
		// mid.Authenticate(a),
		// mid.Authorize(log, auth.RoleAdmin),
	)
	app.Handle(
		http.MethodGet,
		"/liveness",
		check.readiness,
	)

	return app
}
