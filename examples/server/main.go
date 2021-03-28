package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/go-git/go-git/v5"
	"github.com/ipfs/go-merkledag/dagutils"
	"github.com/multiverse-vcs/go-git-ipfs/server"
	"github.com/multiverse-vcs/go-git-ipfs/storage"
	"github.com/multiverse-vcs/go-git-ipfs/unixfs"
)

const bindAddr = "localhost:3030"

// this example clones a repository and starts a git http server
// that you can use to push and clone repositories to and from.
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// create an ephemeral memory backed dag service
	ds := dagutils.NewMemoryDagService()

	// create a unixfs to store git objects
	fs, err := unixfs.New(ctx, ds)
	if err != nil {
		log.Fatal(err)
	}

	// clone this repo as an example
	opts := git.CloneOptions{
		URL: "https://github.com/multiverse-vcs/go-git-ipfs",
	}

	// clone the repo into the unixfs
	_, err = git.CloneContext(ctx, storage.NewStorage(fs), nil, &opts)
	if err != nil {
		log.Fatal(err)
	}

	// add the final unixfs node
	id, err := fs.Save()
	if err != nil {
		log.Fatal(err)
	}

	// you can now clone from this remote
	fmt.Printf("git clone http://%s/%s\n", bindAddr, id.String())

	// create the server instance
	server := server.NewServer(ds)

	// start the http server
	log.Fatal(http.ListenAndServe(bindAddr, server))
}
