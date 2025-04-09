package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/go-git/go-billy/v5/osfs"
	"github.com/golang-cz/devslog"
	"s.nhnn.dev/olshamb/gitaway/capabilities"
	"s.nhnn.dev/olshamb/gitaway/middlewares"
)

func main() {
	logger := slog.New(devslog.NewHandler(os.Stderr, nil))
	slog.SetDefault(logger)
	r := http.NewServeMux()
	info := capabilities.Info{
		Version:      0,
		Capabilities: []capabilities.Capability{},
	}
	reposFs := osfs.New("./git-repos")
	info.Capabilities = capabilities.AddGitCapabilities(r, info, reposFs)
	info.Capabilities = capabilities.AddInfoCapability(r, info)
	http.ListenAndServe("127.0.0.1:3000", middlewares.LoggingMiddleware(r))
}
