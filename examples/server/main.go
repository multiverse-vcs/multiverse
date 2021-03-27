package main

import (
	"context"
	"fmt"
	"net/http"
	"log"

	"github.com/go-git/go-git/v5"
	"github.com/ipfs/go-merkledag/dagutils"
	"github.com/multiverse-vcs/go-git-ipfs/unixfs"
	"github.com/multiverse-vcs/go-git-ipfs/server"
	"github.com/multiverse-vcs/go-git-ipfs/storage"
)

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

	// initialize an empty repo that we can push to
	if _, err = git.Init(storage.NewStorage(fs), nil); err != nil {
		log.Fatal(err)
	}

	// get the final unixfs node
	node, err := fs.Node()
	if err != nil {
		log.Fatal(err)
	}

	// add the empty repo to the dag
	if err := ds.Add(ctx, node); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Push new repositories to this remote")
	fmt.Printf("http://localhost:3030/%s\n", node.Cid().String())

	// create the server instance
	server := server.NewServer(ctx, ds)

	// start the http server
	log.Fatal(http.ListenAndServe(":3030", server))
}
