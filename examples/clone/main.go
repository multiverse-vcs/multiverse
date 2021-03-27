package main

import (
	"context"
	"fmt"
	"log"

	"github.com/go-git/go-git/v5"
	"github.com/ipfs/go-merkledag/dagutils"
	"github.com/multiverse-vcs/go-git-ipfs/storage"
	"github.com/multiverse-vcs/go-git-ipfs/unixfs"
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

	// clone this repo as an example
	opts := git.CloneOptions{
		URL: "https://github.com/multiverse-vcs/go-git-ipfs",
	}

	// clone the repo into the unixfs
	repo, err := git.CloneContext(ctx, storage.NewStorage(fs), nil, &opts)
	if err != nil {
		log.Fatal(err)
	}

	// get the final unixfs node
	node, err := fs.Node()
	if err != nil {
		log.Fatal(err)
	}

	// add the node to the dag service
	if err := ds.Add(ctx, node); err != nil {
		log.Fatal(err)
	}

	// print the cid so we can share it with others
	// NOTE: this example uses an ephemeral dag service
	fmt.Println(node.Cid().String())

	// get the repo HEAD
	ref, err := repo.Head()
	if err != nil {
		log.Fatal(err)
	}

	// get the commit object
	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(commit)
}
