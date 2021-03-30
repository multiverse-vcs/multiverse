package git

import (
	"net/http"

	"github.com/go-git/go-git/v5/plumbing/protocol/packp"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/server"
	cid "github.com/ipfs/go-cid"
	"github.com/gorilla/mux"

	"github.com/multiverse-vcs/go-git-ipfs/internal/database"
)

// UploadPack sends a packfile containing requested references.
func (s *Git) UploadPack(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	params := mux.Vars(req)
	username := params["user"]
	reponame := params["repo"]
	
	var user database.User
	if err := user.FindByUsername(s.DB, username); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var repo database.Repo
	if err := repo.FindByNameAndUserID(s.DB, reponame, user.ID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, err := cid.Decode(repo.CID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	loader := NewLoader(ctx, s.Node.DAG, id)
	server := server.NewServer(loader)

	ep, err := transport.NewEndpoint(req.RequestURI)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sess, err := server.NewUploadPackSession(ep, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sessreq := packp.NewUploadPackRequest()
	if err := sessreq.Decode(req.Body); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sessres, err := sess.UploadPack(ctx, sessreq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Cache-Control", "no-cache")
	w.Header().Add("Content-Type", "application/x-git-upload-pack-result")
	w.WriteHeader(http.StatusOK)
	sessres.Encode(w)
}
