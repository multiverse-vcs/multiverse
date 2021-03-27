package server

import (
	"context"
	"net/http"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/format/pktline"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/server"
	ipld "github.com/ipfs/go-ipld-format"
)

type Server struct {
	server transport.Transport
}

func NewServer(ctx context.Context, ds ipld.DAGService) *Server {
	loader := NewLoader(ctx, ds)
	server := server.NewServer(loader)
	return &Server{server}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Cache-Control", "no-cache")

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

func (s *Server) UploadPack(w http.ResponseWriter, req *http.Request) {
	ep, err := transport.NewEndpoint(req.URL.String())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sess, err := s.server.NewUploadPackSession(ep, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sessreq := packp.NewUploadPackRequest()
	if err := sessreq.Decode(req.Body); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sessres, err := sess.UploadPack(req.Context(), sessreq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/x-git-upload-pack-result")
	sessres.Encode(w)
}

func (s *Server) ReceivePack(w http.ResponseWriter, req *http.Request) {
	ep, err := transport.NewEndpoint(req.URL.String())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sess, err := s.server.NewReceivePackSession(ep, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sessreq := packp.NewReferenceUpdateRequest()
	if err := sessreq.Decode(req.Body); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sessres, err := sess.ReceivePack(req.Context(), sessreq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/x-git-receive-pack-result")
	sessres.Encode(w)
}

func (s *Server) AdvertisedReferences(w http.ResponseWriter, req *http.Request) {
	ep, err := transport.NewEndpoint(req.URL.String())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sess, err := s.server.NewReceivePackSession(ep, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	refs, err := sess.AdvertisedReferences()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch req.URL.Query().Get("service") {
	case "git-receive-pack":
		w.Header().Add("Content-Type", "application/x-git-receive-pack-advertisement")

		enc := pktline.NewEncoder(w)
		enc.EncodeString("# service=git-receive-pack\n")
		enc.Flush()

		refs.Encode(w)
	case "git-upload-pack":
		w.Header().Add("Content-Type", "application/x-git-upload-pack-advertisement")

		enc := pktline.NewEncoder(w)
		enc.EncodeString("# service=git-upload-pack\n")
		enc.Flush()

		refs.Encode(w)
	default:
		http.NotFound(w, req)
	}
}
