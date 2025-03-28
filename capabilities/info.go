package capabilities

import (
	"encoding/json"
	"net/http"
)

type Capability interface {
	Id() string
}
type CapabilityWithHandler interface {
	Path() string
	Handler(w http.ResponseWriter, r *http.Request)
}

func AddSingleRouteCapability(r *http.ServeMux, capability CapabilityWithHandler) {
	r.HandleFunc(capability.Path(), capability.Handler)
}

type Info struct {
	Version      uint8 `json:"version"`
	Capabilities []Capability
}

type InfoCapability struct {
	CapabilityId   string `json:"id"`
	CapabilityPath string `json:"path"`
	info           *Info
}

func AddInfoCapability(r *http.ServeMux, info Info) []Capability {
	capability := InfoCapability{
		CapabilityId:   "core.info",
		CapabilityPath: "GET /info",
		info:           &info,
	}
	info.Capabilities = append(info.Capabilities, capability)
	AddSingleRouteCapability(r, capability)
	return info.Capabilities
}

func (i InfoCapability) Id() string {
	return i.CapabilityId
}

func (i InfoCapability) Path() string {
	return i.CapabilityPath
}

func (i InfoCapability) Handler(w http.ResponseWriter, r *http.Request) {
	encoder := json.NewEncoder(w)
	err := encoder.Encode(&i.info)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error.internal"))
	}
}
