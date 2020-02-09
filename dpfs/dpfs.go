package dpfs

import (
	"sync"

	"github.com/billziss-gh/cgofuse/fuse"
	duplicacy "github.com/gilbertchen/duplicacy/src"
)

// Dpfs is the Duplicacy filesystem type
type Dpfs struct {
	fuse.FileSystemBase
	config     *duplicacy.Config
	storage    duplicacy.Storage
	root       string
	snapshotid string
	revision   int
	password   string
	preference *duplicacy.Preference
	repository string
	files      sync.Map
}

func NewDuplicacyfs() *Dpfs {
	self := Dpfs{}
	return &self
}
