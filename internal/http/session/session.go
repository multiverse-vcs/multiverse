package session

import (
	"net/http"

	"gorm.io/gorm"

	"github.com/multiverse-vcs/go-git-ipfs/internal/database"
)

// CookieName is the name of the http cookie.
const CookieName = "session"

// Get returns the session from the cookie.
func Get(req *http.Request, db *gorm.DB) (*database.Session, error) {
	cookie, err := req.Cookie(CookieName)
	if err != nil {
		return nil, err
	}

	var sess database.Session
	if err := sess.Find(db, cookie.Value); err != nil {
		return nil, err
	}

	return &sess, nil
}

// Set writes the session cookie.
func Set(w http.ResponseWriter, db *gorm.DB, sess *database.Session) error {
	if err := sess.Create(db); err != nil {
		return err
	}

	cookie := http.Cookie{
		Name: CookieName,
		Value: sess.ID,
		HttpOnly: true,
	}

	http.SetCookie(w, &cookie)
	return nil
}

// Clear removes the session cookie.
func Clear(w http.ResponseWriter) {
	cookie := http.Cookie{
		Name: CookieName,
		MaxAge: -1,
	}

	http.SetCookie(w, &cookie)
}