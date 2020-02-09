package main

import (
	"os"

	"github.com/billziss-gh/cgofuse/fuse"
	log "github.com/sirupsen/logrus"
	"gitlab.com/andrewheberle/duplicacy-fuse/dpfs"
)

func main() {
	if len(os.Args) <= 1 {
		log.Fatal("missing mountpoint")
	}

	duplicacyfs := dpfs.NewDuplicacyfs()
	host := fuse.NewFileSystemHost(duplicacyfs)
	host.Mount("", os.Args[1:])
}
