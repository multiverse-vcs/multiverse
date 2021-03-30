package auth

import (
	"net/http"

	"gorm.io/gorm"

	"github.com/multiverse-vcs/go-git-ipfs/internal/database"
	"github.com/multiverse-vcs/go-git-ipfs/internal/http/session"
	"github.com/multiverse-vcs/go-git-ipfs/internal/view"
)

func (s *Auth) SignUp(w http.ResponseWriter, req *http.Request) {
	view.Render(w, "sign_up.html", nil)
}

func (s *Auth) SignUpForm(w http.ResponseWriter, req *http.Request) {
	username := req.FormValue("username")
	email := req.FormValue("email")
	password := req.FormValue("password")

	var user database.User
	if err := user.FindByEmailOrUsername(s.DB, email, username); err != gorm.ErrRecordNotFound {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user = database.User{
		Username: username,
		Email:    email,
		Password: password,
	}

	if err := user.Create(s.DB); err != nil {
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

	http.Redirect(w, req, "/", http.StatusSeeOther)
}
