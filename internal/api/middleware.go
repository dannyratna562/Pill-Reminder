package api

import (
	"log/slog"
	"net/http"
	"time"
)

// requestLogger logs each request's method, path, status, and duration via
// the given slog.Logger.
//
// TODO(story-03): this is also where RequireFamilyLink should live — a
// shared middleware validating the requester is linked to the parent_id in
// question via the family_link table, so every write endpoint from
// story-03 onward gets that check without duplicating it per handler.
func requestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(sw, r)
			logger.Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", sw.status,
				"duration_ms", time.Since(start).Milliseconds(),
			)
		})
	}
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (sw *statusWriter) WriteHeader(status int) {
	sw.status = status
	sw.ResponseWriter.WriteHeader(status)
}
