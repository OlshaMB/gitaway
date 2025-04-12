package repository

import (
	"log"
	"net/http"

	githttp "github.com/go-git/go-git/v6/plumbing/http"
	"github.com/go-git/go-git/v6/plumbing/transport"
)

func TransportRoutes(loader transport.Loader, errlogger *log.Logger) http.Handler {
	return &githttp.Handler{
		Loader:   loader,
		ErrorLog: errlogger,
	}
}
