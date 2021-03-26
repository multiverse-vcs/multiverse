package main

import (
	"context"
	"fmt"
	"log"

	"github.com/go-git/go-git/v5"
	cid "github.com/ipfs/go-cid"
	ipld "github.com/ipfs/go-ipld-format"
	ipfs "github.com/multiverse-vcs/go-git-ipfs"
	"github.com/multiverse-vcs/go-git-ipfs/unixfs"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// this example requires a dag service with a libp2p host and router
	// the code below is just for demonstration purposes
	//
	// for a guide on how to create a connected dag service
	// see https://github.com/hsanjuan/ipfs-lite
	// or https://github.com/ipfs/go-ipfs/tree/master/docs/examples/go-ipfs-as-a-library
	var ds ipld.DAGService

	// decode the CID of the repo we want to open
	id, err := cid.Decode("QmdQniaf8v3f2xMxB5AUcJJg5Jj8vJJNTpfs5cJpk2KkB1")
	if err != nil {
		log.Fatal(err)
	}

	// load the unixfs from the CID we just decoded
	fs, err := unixfs.Load(ctx, ds, id)
	if err != nil {
		log.Fatal(err)
	}

	// open the git repo using the unixfs storage
	repo, err := git.Open(ipfs.NewStorage(fs), nil)
	if err != nil {
		log.Fatal(err)
	}

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
