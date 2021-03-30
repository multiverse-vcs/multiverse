package home

import (
	"net/http"

	"github.com/multiverse-vcs/go-git-ipfs/internal/database"
	"github.com/multiverse-vcs/go-git-ipfs/internal/http/session"
	"github.com/multiverse-vcs/go-git-ipfs/internal/view"
)

func (s *Home) Read(w http.ResponseWriter, req *http.Request) {
	data := make(map[string]interface{})

	sess, err := session.Get(req, s.DB)
	if err == nil {
		data["Session"] = sess
	}

	var repos []database.Repo
	if err := s.DB.Limit(10).Order("updated_at desc").Preload("User").Find(&repos).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data["Repos"] = repos
	view.Render(w, "home.html", data)
}
