package server

import (
	"net/http"
	"strings"

	ipld "github.com/ipfs/go-ipld-format"
)

// Server is a git http backend.
type Server struct {
	ds ipld.DAGService
}

// NewServer returns a new server that uses the given dag service.
func NewServer(ds ipld.DAGService) *Server {
	return &Server{ds}
}

// UploadPack sends a packfile containing requested references.
func (s *Server) UploadPack(w http.ResponseWriter, req *http.Request) {
	session := NewSession(req.Context(), s.ds)
	session.UploadPack(w, req)
}

// ReceivePack updates a repository with a packfile and replies with a status.
func (s *Server) ReceivePack(w http.ResponseWriter, req *http.Request) {
	session := NewSession(req.Context(), s.ds)
	session.ReceivePack(w, req)
}

// AdvertisedReferences retrieves the advertised references for a repository.
func (s *Server) AdvertisedReferences(w http.ResponseWriter, req *http.Request) {
	session := NewSession(req.Context(), s.ds)
	session.AdvertisedReferences(w, req)
}

// ServeHTTP is the entry point for all git http requests.
func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch {
	case strings.HasSuffix(req.URL.Path, "/git-upload-pack"):
		s.UploadPack(w, req)
	case strings.HasSuffix(req.URL.Path, "/git-receive-pack"):
		s.ReceivePack(w, req)
	case strings.HasSuffix(req.URL.Path, "/info/refs"):
		s.AdvertisedReferences(w, req)
	default:
		http.NotFound(w, req)
	}
}
