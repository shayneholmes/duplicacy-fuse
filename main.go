package main

import (
	"os"
	"path"
	"strings"
	"sync"

	"github.com/billziss-gh/cgofuse/fuse"
	duplicacy "github.com/gilbertchen/duplicacy/src"
	log "github.com/sirupsen/logrus"
)

const (
	contents = "hello, world\n"
)

// Dpfs is the Duplicacy filesystem type
type Dpfs struct {
	fuse.FileSystemBase
	config     *duplicacy.Config
	storage    duplicacy.Storage
	lock       sync.RWMutex
	root       string
	password   string
	preference *duplicacy.Preference
	repository string
	files      sync.Map
}

func NewDuplicacyfs() *Dpfs {
	self := Dpfs{}
	return &self
}

func (self *Dpfs) Open(path string, flags int) (errc int, fh uint64) {
	switch path {
	case "/hello":
		return 0, 0
	default:
		return -fuse.ENOENT, ^uint64(0)
	}
}

func (self *Dpfs) Read(path string, buff []byte, ofst int64, fh uint64) (n int) {
	endofst := ofst + int64(len(buff))
	if endofst > int64(len(contents)) {
		endofst = int64(len(contents))
	}
	if endofst < ofst {
		return 0
	}
	n = copy(buff, contents[ofst:endofst])
	return
}

func (self *Dpfs) snapshotPath(p string) string {
	self.lock.RLock()
	defer self.lock.RUnlock()

	return strings.TrimSuffix(path.Join(self.root, p), "/")
}

func (self *Dpfs) revision(p string) (revision string) {
	self.lock.RLock()
	defer self.lock.RUnlock()

	slice := strings.Split(p, "/")

	if len(slice) < 2 {
		return ""
	}

	if self.root == "snapshots" {
		if len(slice) < 3 {
			return ""
		} else {
			return slice[2]
		}
	}

	return slice[1]
}

func main() {
	// debug while in dev
	log.SetLevel(log.DebugLevel)

	if len(os.Args) <= 1 {
		log.Fatal("missing mountpoint")
	}

	duplicacyfs := NewDuplicacyfs()
	host := fuse.NewFileSystemHost(duplicacyfs)
	host.Mount("", os.Args[1:])
}
