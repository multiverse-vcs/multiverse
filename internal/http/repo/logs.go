package repo

import (
	"net/http"
	"strconv"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/gorilla/mux"

	"github.com/multiverse-vcs/go-git-ipfs/internal/database"
	"github.com/multiverse-vcs/go-git-ipfs/internal/http/session"
	"github.com/multiverse-vcs/go-git-ipfs/internal/gitutil"
	"github.com/multiverse-vcs/go-git-ipfs/internal/view"
)

func (s *Repo) Logs(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	data := make(map[string]interface{})

	sess, err := session.Get(req, s.DB)
	if err == nil {
		data["Session"] = sess
	}

	params := mux.Vars(req)
	username := params["user"]
	reponame := params["repo"]
	offset := req.URL.Query().Get("offset")

	if offset == "" {
		offset = "0"
	}

	offsetnum, err := strconv.ParseInt(offset, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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

	data["User"] = user
	data["Repo"] = repo
	data["Tab"] = RepoLogsTab

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

	logs, err := gitutil.Logs(git, head, int(offsetnum), RepoLogsPerPage)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data["Commits"] = logs
	data["Next"] = offsetnum + RepoLogsPerPage
	data["Prev"] = offsetnum - RepoLogsPerPage
	view.Render(w, "repo.html", data)
}
