package api

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

func NewRouter(pairingHandler *PairingHandler, logger *slog.Logger) http.Handler {
	r := chi.NewRouter()
	r.Use(chimw.Recoverer)
	r.Use(requestLogger(logger))

	r.Route("/v1", func(r chi.Router) {
		r.Post("/pairing-codes", pairingHandler.RequestCode)
		r.Post("/family-links", pairingHandler.Redeem)
	})

	return r
}
