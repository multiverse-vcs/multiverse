package git

import (
	// "io"
	// "os"
	"net/http"

	"github.com/go-git/go-git/v5/plumbing/protocol/packp"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/server"
	cid "github.com/ipfs/go-cid"
	"github.com/gorilla/mux"

	"github.com/multiverse-vcs/go-git-ipfs/internal/database"
)

// ReceivePack updates a repository with a packfile and replies with a status.
func (s *Git) ReceivePack(w http.ResponseWriter, req *http.Request) {
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

	// acquire a pinlock so GC doesn't wipe out changes
	defer s.Node.Blockstore.PinLock().Unlock()

	ep, err := transport.NewEndpoint(req.RequestURI)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sess, err := server.NewReceivePackSession(ep, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// io.Copy(os.Stdout, req.Body)
	// w.Header().Add("Cache-Control", "no-cache")
	// w.Header().Add("Content-Type", "application/x-git-receive-pack-result")
	// return

	sessreq := packp.NewReferenceUpdateRequest()
	if err := sessreq.Decode(req.Body); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sessres, err := sess.ReceivePack(ctx, sessreq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	node, err := loader.Node()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := s.Node.Pinning.Pin(ctx, node, true); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	repo.CID = node.Cid().String()
	if err := repo.UpdateCID(s.DB); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Cache-Control", "no-cache")
	w.Header().Add("Content-Type", "application/x-git-receive-pack-result")
	w.WriteHeader(http.StatusOK)

	sessres.Encode(w)
}
