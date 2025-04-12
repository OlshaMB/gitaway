package repository

import (
	"log/slog"
	"net/http"

	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v6/plumbing/transport"
)

var (
	DefaultLoader = transport.NewFilesystemLoader(osfs.New("./git-repos"), false)
)

func RepositoryRoutes(loader transport.Loader) http.Handler {
	m := http.NewServeMux()
	if loader == nil {
		loader = DefaultLoader
	}
	m.Handle("/", TransportRoutes(
		loader,
		slog.NewLogLogger(slog.Default().Handler().WithGroup("git-transport"), slog.LevelError),
	))
	return m
}
