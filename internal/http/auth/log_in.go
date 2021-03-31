package auth

import (
	"net/http"

	"golang.org/x/crypto/bcrypt"

	"github.com/multiverse-vcs/go-git-ipfs/internal/database"
	"github.com/multiverse-vcs/go-git-ipfs/internal/http/session"
	"github.com/multiverse-vcs/go-git-ipfs/internal/view"
)

func (s *Auth) LogIn(w http.ResponseWriter, req *http.Request) {
	view.Render(w, "log_in.html", nil)
}

func (s *Auth) LogInForm(w http.ResponseWriter, req *http.Request) {
	username := req.FormValue("username")
	password := req.FormValue("password")

	var user database.User
	if err := user.FindByUsername(s.DB, username); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sess := database.Session{
		UserID: user.ID,
	}

	if err := session.Set(w, s.DB, &sess); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, req, "/"+user.Username, http.StatusSeeOther)
}
