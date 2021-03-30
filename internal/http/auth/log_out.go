package auth

import (
	"net/http"

	"github.com/multiverse-vcs/go-git-ipfs/internal/http/session"
)

func (s *Auth) LogOut(w http.ResponseWriter, req *http.Request) {
	session.Clear(w)
	http.Redirect(w, req, "/", http.StatusSeeOther)
}
