package dpfs

import (
	"sync"

	"github.com/billziss-gh/cgofuse/fuse"
	duplicacy "github.com/gilbertchen/duplicacy/src"
)

// Dpfs is the Duplicacy filesystem type. This type satisfies the fuse.FileSystemInterface interace
type Dpfs struct {
	fuse.FileSystemBase
	config          *duplicacy.Config
	storage         duplicacy.Storage
	chunkDownloader *duplicacy.ChunkDownloader
	root            string
	snapshotid      string
	revision        int
	password        string
	preference      *duplicacy.Preference
	repository      string
	mu              sync.Mutex
	cache           DpfsKvStore
}

// Nicer names for fuse errors/return codes
const (
	NotImplemented        = -fuse.ENOSYS
	NoSuchFileOrDirectory = -fuse.ENOENT
	IOError               = -fuse.EIO
	IsDirectory           = -fuse.EISDIR
	NotDirectory          = -fuse.ENOTDIR
)

// NewDuplicacyfs creates an initial Dpfs struct
func NewDuplicacyfs() *Dpfs {
	self := Dpfs{}
	return &self
}
