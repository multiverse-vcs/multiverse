package user

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/multiverse-vcs/go-git-ipfs/internal/database"
	"github.com/multiverse-vcs/go-git-ipfs/internal/http/session"
	"github.com/multiverse-vcs/go-git-ipfs/internal/view"
)

func (s *User) Read(w http.ResponseWriter, req *http.Request) {
	data := make(map[string]interface{})

	sess, err := session.Get(req, s.DB)
	if err == nil {
		data["Session"] = sess
	}

	params := mux.Vars(req)
	username := params["user"]

	var user database.User
	if err := user.FindByUsername(s.DB, username); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var repos []database.Repo
	if err := s.DB.Find(&repos, "user_id = ?", user.ID).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data["User"] = user
	data["Repos"] = repos
	view.Render(w, "user.html", data)
}
