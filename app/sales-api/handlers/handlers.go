// Package handlers contains the full set of handler functions and routes
// supported by the web api.
package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/dimfeld/httptreemux/v5"
)

// API constructs an http.Handler with all application routes defined.
func API(build string, shutdown chan os.Signal, log *log.Logger) *httptreemux.ContextMux {
	tm := httptreemux.NewContextMux()

	h := func(w http.ResponseWriter, r *http.Request) {
		status := struct {
			Status string
		}{
			Status: "OK",
		}

		json.NewEncoder(w).Encode(status)
	}

	tm.Handle(http.MethodGet, "/test", h)

	return tm
}
