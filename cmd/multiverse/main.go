package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/multiverse-vcs/go-git-ipfs/internal/core"
	"github.com/multiverse-vcs/go-git-ipfs/internal/http"
)

const banner = `
  __  __       _ _   _                         
 |  \/  |_   _| | |_(_)_   _____ _ __ ___  ___ 
 | |\/| | | | | | __| \ \ / / _ \ '__/ __|/ _ \
 | |  | | |_| | | |_| |\ V /  __/ |  \__ \  __/
 |_|  |_|\__,_|_|\__|_| \_/ \___|_|  |___/\___|
                                               
`

func main() {
	server, err := core.NewServer(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	web := http.NewServer(server)
	go web.ListenAndServe()

	fmt.Print(banner)
	fmt.Println("your peer id is", server.Node.Identity)
	fmt.Println("web server listening on", web.Addr)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Println("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	web.Shutdown(ctx)
	server.Node.Close()
}
