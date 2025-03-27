package middlewares

import (
	"log/slog"
	"math/rand/v2"
	"net/http"
	"strconv"
)

func LoggingMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := rand.Uint64()
		paid := r.Pattern + " " + strconv.FormatUint(id, 10)
		slog.Info(paid, "Query", r.URL.Query())
		h.ServeHTTP(w, r)
		slog.Info(paid)
	})
}
