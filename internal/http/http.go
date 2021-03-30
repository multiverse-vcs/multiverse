package http

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/multiverse-vcs/go-git-ipfs/internal/core"
	"github.com/multiverse-vcs/go-git-ipfs/internal/http/auth"
	"github.com/multiverse-vcs/go-git-ipfs/internal/http/git"
	"github.com/multiverse-vcs/go-git-ipfs/internal/http/home"
	"github.com/multiverse-vcs/go-git-ipfs/internal/http/repo"
	"github.com/multiverse-vcs/go-git-ipfs/internal/http/user"
	"github.com/multiverse-vcs/go-git-ipfs/internal/view"
)

// ListenAndServe starts accepting http connections.
func ListenAndServe(server *core.Server) error {
	// TODO disable
	view.Development = true

	router := mux.NewRouter()

	auth := (*auth.Auth)(server)
	git := (*git.Git)(server)
	home := (*home.Home)(server)
	repo := (*repo.Repo)(server)
	user := (*user.User)(server)

	static := http.FileServer(http.Dir("web/public"))
	static = http.StripPrefix("/public/", static)

	router.HandleFunc("/", home.Read).Methods(http.MethodGet)
	router.HandleFunc("/_create_repo", repo.Create).Methods(http.MethodGet)
	router.HandleFunc("/_create_repo", repo.CreateForm).Methods(http.MethodPost)
	router.HandleFunc("/_sign_up", auth.SignUp).Methods(http.MethodGet)
	router.HandleFunc("/_sign_up", auth.SignUpForm).Methods(http.MethodPost)
	router.HandleFunc("/_log_in", auth.LogIn).Methods(http.MethodGet)
	router.HandleFunc("/_log_in", auth.LogInForm).Methods(http.MethodPost)
	router.HandleFunc("/_log_out", auth.LogOut).Methods(http.MethodGet)

	router.HandleFunc("/{user}", user.Read).Methods(http.MethodGet)
	router.HandleFunc("/{user}/{repo}", repo.Read).Methods(http.MethodGet)
	router.HandleFunc("/{user}/{repo}/tree/{refpath:.*}", repo.Tree).Methods(http.MethodGet)
	router.HandleFunc("/{user}/{repo}/logs", repo.Logs).Methods(http.MethodGet)
	router.HandleFunc("/{user}/{repo}/refs", repo.Refs).Methods(http.MethodGet)
	router.HandleFunc("/{user}/{repo}/git-upload-pack", git.UploadPack).Methods(http.MethodPost)
	router.HandleFunc("/{user}/{repo}/git-receive-pack", git.ReceivePack).Methods(http.MethodPost)
	router.HandleFunc("/{user}/{repo}/info/refs", git.AdvertisedReferences).Methods(http.MethodGet)

	l := logger{router}

	http.Handle("/public/", static)
	http.Handle("/", l)
	return http.ListenAndServe(":3000", nil)
}

type logger struct {
	router http.Handler
}

func (l logger) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	fmt.Println(req.RequestURI)
	l.router.ServeHTTP(w, req)
}