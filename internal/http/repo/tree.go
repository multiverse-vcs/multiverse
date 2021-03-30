package repo

import (
	"net/http"

	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/gorilla/mux"

	"github.com/multiverse-vcs/go-git-ipfs/internal/database"
	"github.com/multiverse-vcs/go-git-ipfs/internal/http/session"
	"github.com/multiverse-vcs/go-git-ipfs/internal/gitutil"
	"github.com/multiverse-vcs/go-git-ipfs/internal/view"
)

func (s *Repo) Tree(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	data := make(map[string]interface{})

	sess, err := session.Get(req, s.DB)
	if err == nil {
		data["Session"] = sess
	}

	params := mux.Vars(req)
	username := params["user"]
	reponame := params["repo"]
	refpath := params["refpath"]

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

	ref, path, err := gitutil.RefPath(git, refpath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	obj, err := gitutil.Find(git, ref, path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch o := obj.(type) {
	case *object.Tree:
		data["Tree"] = o
	case *object.Blob:
		data["Blob"] = o
	}

	data["User"] = user
	data["Repo"] = repo
	data["Ref"] = ref.Name().String()
	data["Path"] = path
	data["Tab"] = RepoTreeTab
	view.Render(w, "repo.html", data)
}
