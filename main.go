package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/OlshaMB/gitaway/middlewares"
	"github.com/OlshaMB/gitaway/routes/repository"
	"github.com/golang-cz/devslog"
)

func main() {
	logger := slog.New(devslog.NewHandler(os.Stderr, nil))
	slog.SetDefault(logger)
	r := http.NewServeMux()
	r.Handle("/{name}/", repository.RepositoryRoutes(nil))
	http.ListenAndServe("127.0.0.1:3000", middlewares.LoggingMiddleware(r))
}
