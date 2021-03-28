package server

import (
	"context"
	"net/http"

	"github.com/go-git/go-git/v5/plumbing/format/pktline"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/server"
	ipld "github.com/ipfs/go-ipld-format"
)

// Session is a git http request.
type Session struct {
	loader *Loader
	server transport.Transport
}

// NewSession returns a new session.
func NewSession(ctx context.Context, ds ipld.DAGService) *Session {
	loader := NewLoader(ctx, ds)
	server := server.NewServer(loader)
	return &Session{loader, server}
}

// UploadPack sends a packfile containing requested references.
func (s *Session) UploadPack(w http.ResponseWriter, req *http.Request) {
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

// ReceivePack updates a repository with a packfile and replies with a status.
func (s *Session) ReceivePack(w http.ResponseWriter, req *http.Request) {
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

	if _, err := s.loader.Save(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/x-git-receive-pack-result")
	sessres.Encode(w)
}

// AdvertisedReferences retrieves the advertised references for a repository.
func (s *Session) AdvertisedReferences(w http.ResponseWriter, req *http.Request) {
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

	var typ, svc string
	switch req.URL.Query().Get("service") {
	case transport.UploadPackServiceName:
		typ = "application/x-git-upload-pack-advertisement"
		svc = "# service=git-upload-pack\n"
	case transport.ReceivePackServiceName:
		typ = "application/x-git-receive-pack-advertisement"
		svc = "# service=git-receive-pack\n"
	default:
		http.NotFound(w, req)
		return
	}

	w.Header().Add("Content-Type", typ)
	w.Header().Add("Cache-Control", "no-cache")

	enc := pktline.NewEncoder(w)
	enc.EncodeString(svc)
	enc.Flush()

	refs.Encode(w)
}
