package capabilities

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/server"
)

var (
	ErrorInvalidService = errors.New("Invalid service")
)

type GitTransportCapability struct {
	CapabilityId string `json:"id"`
	srv          transport.Transport
}

func AddGitCapabilities(r *http.ServeMux, info Info) []Capability {
	capabilities := info.Capabilities
	loader := server.NewFilesystemLoader(osfs.New("./git-repos"))
	srv := server.NewServer(loader)
	gitRemoteHttpsCapability := GitTransportCapability{
		CapabilityId: "git.remote.https",
		srv:          srv,
	}
	capabilities = append(capabilities, gitRemoteHttpsCapability)
	gr := http.NewServeMux()
	gr.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	})
	gr.HandleFunc("GET /{name}/info/refs", gitRemoteHttpsCapability.RefsHandler)
	gr.HandleFunc("POST /{name}/git-upload-pack", gitRemoteHttpsCapability.UploadPackHandler)
	gr.HandleFunc("POST /{name}/git-receive-pack", gitRemoteHttpsCapability.ReceivePackHandler)
	r.Handle("/repos/", http.StripPrefix("/repos", gr))
	return capabilities
}

func (g GitTransportCapability) Id() string {
	return g.CapabilityId
}

func (g GitTransportCapability) newSession(service string, repoName string) (transport.Session, error) {
	var session transport.Session
	endpoint, err := transport.NewEndpoint("/" + repoName)
	if err != nil {
		return nil, err
	}
	if service == "receive-pack" {
		session, err = g.srv.NewReceivePackSession(endpoint, nil)
	} else if service == "upload-pack" {
		session, err = g.srv.NewUploadPackSession(endpoint, nil)
	} else {
		slog.Info(service)
		return nil, ErrorInvalidService
	}
	if err != nil {
		return nil, err
	}
	return session, nil
}
func (g GitTransportCapability) RefsHandler(w http.ResponseWriter, r *http.Request) {
	service := strings.TrimPrefix(r.URL.Query().Get("service"), "git-")
	slog.Info(service)
	session, err := g.newSession(service, r.PathValue("name"))
	if err != nil {
		slog.Error("error", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error.internal"))
		return
	}
	defer session.Close()
	refs, err := session.AdvertisedReferences()
	if err != nil {
		slog.Error("error", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error.internal"))
		return
	}

	w.Header().Set("Content-Type", fmt.Sprintf("application/x-git-%s-advertisement", service))
	w.WriteHeader(http.StatusOK)
	refs.Prefix = [][]byte{
		[]byte(fmt.Sprintf("# service=git-%s", service)),
		nil,
	}
	// if refs.IsEmpty() {
	// 	encoder := pktline.NewEncoder(w)
	// 	encoder.Encode([]byte(fmt.Sprintf("# service=git-%s\n", service)))
	// 	encoder.Encode(nil)
	// 	return
	// }
	slog.Info("a", "a", len(refs.References))
	if err := refs.Encode(w); err != nil {
		slog.Error("Something went wrong while encoding")
	}
}
func (g GitTransportCapability) UploadPackHandler(w http.ResponseWriter, r *http.Request) {
	session, err := g.newSession("upload-pack", r.PathValue("name"))
	if err != nil {
		slog.Error("error", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error.internal"))
		return
	}
	defer session.Close()
	uploadSession := session.(transport.UploadPackSession)
	uploadRequest := packp.NewUploadPackRequest()
	if err := uploadRequest.Decode(r.Body); err != nil {
		slog.Error("error", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error.internal"))
		return
	}
	uploadResponse, err := uploadSession.UploadPack(r.Context(), uploadRequest)

	if err != nil {
		slog.Error("error", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error.internal"))
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/x-git-upload-pack-result")
	uploadResponse.Encode(w)
}
func (g GitTransportCapability) ReceivePackHandler(w http.ResponseWriter, r *http.Request) {
	session, err := g.newSession("receive-pack", r.PathValue("name"))
	if err != nil {
		slog.Error("error", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error.internal"))
		return
	}
	defer session.Close()
	receiveSession := session.(transport.ReceivePackSession)
	receiveRequest := packp.NewReferenceUpdateRequest()
	if err := receiveRequest.Decode(r.Body); err != nil {
		slog.Error("error", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error.internal"))
		return
	}
	receiveResponse, err := receiveSession.ReceivePack(r.Context(), receiveRequest)

	if err != nil {
		slog.Error("error", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error.internal"))
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/x-git-receive-pack-result")
	receiveResponse.Encode(w)
}
