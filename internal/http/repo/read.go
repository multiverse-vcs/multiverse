package repo

import (
	"net/http"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/gorilla/mux"

	"github.com/multiverse-vcs/go-git-ipfs/internal/database"
	"github.com/multiverse-vcs/go-git-ipfs/internal/http/session"
	"github.com/multiverse-vcs/go-git-ipfs/internal/gitutil"
	"github.com/multiverse-vcs/go-git-ipfs/internal/view"
)

func (s *Repo) Read(w http.ResponseWriter, req *http.Request) {
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

	data["URL"] = req.URL.String()
	data["User"] = user
	data["Repo"] = repo

	git, err := gitutil.Open(ctx, s.Node.DAG, repo.CID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	head, err := git.Head()
	if err == plumbing.ErrReferenceNotFound {
		view.Render(w, "empty_repo.html", data)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	readme, err := gitutil.Readme(git, head)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data["Tab"] = RepoInfoTab
	data["Readme"] = readme
	view.Render(w, "repo.html", data)
}
