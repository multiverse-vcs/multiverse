package repo

import (
	"fmt"
	"net/http"

	"github.com/multiverse-vcs/go-git-ipfs/internal/database"
	"github.com/multiverse-vcs/go-git-ipfs/internal/gitutil"
	"github.com/multiverse-vcs/go-git-ipfs/internal/http/session"
	"github.com/multiverse-vcs/go-git-ipfs/internal/view"
)

func (s *Repo) Create(w http.ResponseWriter, req *http.Request) {
	data := make(map[string]interface{})

	sess, err := session.Get(req, s.DB)
	if err == nil {
		data["Session"] = sess
	}

	view.Render(w, "create_repo.html", data)
}

func (s *Repo) CreateForm(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	name := req.FormValue("name")
	description := req.FormValue("description")

	sess, err := session.Get(req, s.DB)
	if err != nil {
		http.Redirect(w, req, "/_log_in", http.StatusSeeOther)
		return
	}

	var repo database.Repo
	if err := repo.FindByNameAndUserID(s.DB, name, sess.UserID); err == nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// acquire a pinlock so GC doesn't wipe out changes
	defer s.Node.Blockstore.PinLock().Unlock()

	node, err := gitutil.Init(ctx, s.Node.DAG)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := s.Node.Pinning.Pin(ctx, node, true); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	repo = database.Repo{
		Name:          name,
		Description:   description,
		UserID:        sess.UserID,
		CID:           node.Cid().String(),
	}

	if err := repo.Create(s.DB); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	url := fmt.Sprintf("/%s/%s", sess.User.Username, name)
	http.Redirect(w, req, url, http.StatusSeeOther)
}
