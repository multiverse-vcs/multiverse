package git

import (
	"fmt"
	"net/http"

	"github.com/go-git/go-git/v5/plumbing/format/pktline"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/server"
	cid "github.com/ipfs/go-cid"
	"github.com/gorilla/mux"

	"github.com/multiverse-vcs/go-git-ipfs/internal/database"
)

// AdvertisedReferences retrieves the advertised references for a repository.
func (s *Git) AdvertisedReferences(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	params := mux.Vars(req)
	username := params["user"]
	reponame := params["repo"]
	service := req.URL.Query().Get("service")

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

	var sess transport.Session
	var err0 error

	switch service {
	case transport.UploadPackServiceName:
		sess, err0 = server.NewUploadPackSession(ep, nil)
	case transport.ReceivePackServiceName:
		sess, err0 = server.NewReceivePackSession(ep, nil)
	default:
		http.NotFound(w, req)
		return
	}

	if err0 != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	refs, err := sess.AdvertisedReferences()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", fmt.Sprintf("application/x-%s-advertisement", service))
	w.Header().Add("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)

	enc := pktline.NewEncoder(w)
	enc.EncodeString(fmt.Sprintf("# service=%s\n", service))
	enc.Flush()

	refs.Encode(w)
}
