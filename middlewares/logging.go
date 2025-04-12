package middlewares

import (
	"fmt"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"strconv"
	"time"
)

func LoggingMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := rand.Uint64()
		rid := strconv.FormatUint(id, 10) + " " + r.URL.Path
		t := time.Now()
		slog.Info(fmt.Sprintf("-> %s", rid))
		h.ServeHTTP(w, r)
		slog.Info(fmt.Sprintf("<- %s", rid), "t", time.Since(t))
	})
}
