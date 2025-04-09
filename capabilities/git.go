package capabilities

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v6/plumbing/transport"
	"github.com/go-git/go-git/v6/storage"
	"github.com/go-git/go-git/v6/utils/trace"
)

var (
	ErrorInvalidService = errors.New("Invalid service")
)

type GitTransportCapability struct {
	CapabilityId string `json:"id"`
	loader       transport.Loader
}

func AddGitCapabilities(r *http.ServeMux, info Info, fs billy.Filesystem) []Capability {
	capabilities := info.Capabilities
	trace.SetTarget(trace.Packet)
	loader := transport.NewFilesystemLoader(fs, false)

	gitRemoteHttpsCapability := GitTransportCapability{
		CapabilityId: "git.remote.https",
		loader:       loader,
	}
	capabilities = append(capabilities, gitRemoteHttpsCapability)

	gr := http.NewServeMux()
	gr.HandleFunc("GET /{name}/info/refs", gitRemoteHttpsCapability.RefsHandler)
	gr.HandleFunc("POST /{name}/git-upload-pack", gitRemoteHttpsCapability.UploadPackHandler)
	gr.HandleFunc("POST /{name}/git-receive-pack", gitRemoteHttpsCapability.ReceivePackHandler)
	r.Handle("/repos/", http.StripPrefix("/repos", gr))

	return capabilities
}

func (g GitTransportCapability) Id() string {
	return g.CapabilityId
}

type gitTransportResponseWriterCloser struct {
	http.ResponseWriter
}

func (gitTransportResponseWriterCloser) Close() error { return nil }
func (g GitTransportCapability) newStorer(repoName string) (storage.Storer, error) {
	endpoint, err := transport.NewEndpoint("/" + repoName)
	if err != nil {
		return nil, err
	}
	storer, err := g.loader.Load(endpoint)
	if err != nil {
		return nil, err
	}
	return storer, nil
}
func (g GitTransportCapability) RefsHandler(w http.ResponseWriter, r *http.Request) {
	service := r.URL.Query().Get("service")
	slog.Info(service)
	storer, err := g.newStorer(r.PathValue("name"))
	if err != nil {
		slog.Error("error", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error.internal"))
		return
	}
	w.Header().Set("Content-Type", fmt.Sprintf("application/x-%s-advertisement", service))
	w.WriteHeader(http.StatusOK)
	err = transport.AdvertiseReferences(r.Context(), storer, w, transport.Service(service), true)
	if err != nil {
		w.Write([]byte("error.internal"))
		return
	}
}

func (w *gitTransportResponseWriterCloser) Write(p []byte) (n int, err error) {
	if strings.Contains(string(p), "0019") {
		slog.Debug("test")
	}
	return w.ResponseWriter.Write(p)
}

func (g GitTransportCapability) UploadPackHandler(w http.ResponseWriter, r *http.Request) {
	storer, err := g.newStorer(r.PathValue("name"))
	if err != nil {
		slog.Error("error", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error.internal"))
		return
	}
	w.Header().Add("Content-Type", "application/x-git-upload-pack-result")
	w.WriteHeader(http.StatusOK)
	err = transport.UploadPack(r.Context(), storer, r.Body, &gitTransportResponseWriterCloser{w}, &transport.UploadPackOptions{
		// GitProtocol:   protocol.V1.String(),
		AdvertiseRefs: false,
		StatelessRPC:  true,
	})
	if err != nil {
		slog.Error("error", "err", err)
		w.Write([]byte("error.internal"))
	}
}
func (g GitTransportCapability) ReceivePackHandler(w http.ResponseWriter, r *http.Request) {
	storer, err := g.newStorer(r.PathValue("name"))
	if err != nil {
		slog.Error("error", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error.internal"))
		return
	}
	w.Header().Add("Content-Type", "application/x-git-receive-pack-result")
	w.WriteHeader(http.StatusOK)
	// buf := new(strings.Builder)
	// _, err = io.Copy(buf, r.Body)
	// // check errors
	// fmt.Println(buf.String())
	err = transport.ReceivePack(r.Context(), storer, r.Body, &gitTransportResponseWriterCloser{w}, &transport.ReceivePackOptions{
		// GitProtocol:   "version=1",
		AdvertiseRefs: false,
		StatelessRPC:  true,
	})
	// if err != nil {
	// 	slog.Error("error", "err", err)
	// 	w.Write([]byte("error.internal"))
	// }
}
