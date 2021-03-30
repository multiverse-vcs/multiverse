package main

import (
	"context"
	"log"

	"github.com/multiverse-vcs/go-git-ipfs/internal/core"
	"github.com/multiverse-vcs/go-git-ipfs/internal/http"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	server, err := core.NewServer(ctx)
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(http.ListenAndServe(server))
}
