package middleware

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

type contextKey string

const CorrelationIDKey contextKey = "correlation_id"

func CorrelationID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		correlationID := r.Header.Get("X-Correlation-ID")
		if correlationID == "" {
			correlationID = middleware.GetReqID(r.Context())
		}
		ctx := context.WithValue(r.Context(), CorrelationIDKey, correlationID)
		w.Header().Set("X-Correlation-ID", correlationID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
