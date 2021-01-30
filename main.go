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
	// Get fuse-compatible arguments
	var repository, snapshotid, storageName, storagePassword, loglevel, cachedir string
	var revision int
	var debug, all, cleancache bool
	outargs, err := fuse.OptParse(os.Args[1:], "repository=%s storage=%s snapshot=%s revision=%d password=%s loglevel=%s cachedir=%s debug all cleancache", &repository, &storageName, &snapshotid, &revision, &storagePassword, &loglevel, &cachedir, &debug, &all, &cleancache)
	if err != nil {
		log.WithError(err).Fatal("arg error")
	}

	mountpoint := outargs[len(outargs)-1]
	fuseargs := outargs[0 : len(outargs)-1]
	log.Warnf("Calling with mountpoint %v and args: %v", mountpoint, fuseargs)

	host.Mount(mountpoint, fuseargs)
}
