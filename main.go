package main

import (
	"os"
	"sync"

	"github.com/billziss-gh/cgofuse/fuse"
	duplicacy "github.com/gilbertchen/duplicacy/src"
	log "github.com/sirupsen/logrus"
)

// Dpfs is the Duplicacy filesystem type
type Dpfs struct {
	fuse.FileSystemBase
	config     *duplicacy.Config
	storage    duplicacy.Storage
	lock       sync.RWMutex
	root       string
	snapshotid string
	revision   int
	password   string
	preference *duplicacy.Preference
	repository string
	files      sync.Map
	flock      sync.RWMutex
	//	ofiles     map[uint64]node_t
}

func NewDuplicacyfs() *Dpfs {
	self := Dpfs{}
	return &self
}

func main() {
	if len(os.Args) <= 1 {
		log.Fatal("missing mountpoint")
	}

	duplicacyfs := NewDuplicacyfs()
	host := fuse.NewFileSystemHost(duplicacyfs)
	host.Mount("", os.Args[1:])
}
