package repo

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/multiverse-vcs/go-git-ipfs/internal/database"
	"github.com/multiverse-vcs/go-git-ipfs/internal/http/session"
	"github.com/multiverse-vcs/go-git-ipfs/internal/gitutil"
	"github.com/multiverse-vcs/go-git-ipfs/internal/view"
)

func (s *Repo) Refs(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	data := make(map[string]interface{})

	sess, err := session.Get(req, s.DB)
	if err == nil {
		data["Session"] = sess
	}

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

	git, err := gitutil.Open(ctx, s.Node.DAG, repo.CID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	refs, err := gitutil.References(git)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data["User"] = user
	data["Repo"] = repo
	data["Tab"] = RepoRefsTab
	data["References"] = refs
	view.Render(w, "repo.html", data)
}
