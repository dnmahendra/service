package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"math/rand"
	"net/http"

	"github.com/dnmahendra/service/foundation/web"
)

type check struct {
	log *log.Logger
}

func (c check) readiness(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if n := rand.Intn(100); n%2 == 0 {
		return web.NewRequestError(errors.New("trusted error"), http.StatusBadRequest)

		// return web.NewShutdownError("forcing shutdown")
	}

	status := struct {
		Status string
	}{
		Status: "OK",
	}

	return json.NewEncoder(w).Encode(status)
}
