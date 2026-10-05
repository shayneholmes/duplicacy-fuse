package dpfs

import (
	"sync"

	"github.com/billziss-gh/cgofuse/fuse"
	duplicacy "github.com/gilbertchen/duplicacy/src"
	lru "github.com/hashicorp/golang-lru/v2"
	log "github.com/sirupsen/logrus"
)

type revisionCacheKey struct {
	snapshotid string
	revision   int
}

// Dpfs is the Duplicacy filesystem type. This type satisfies the fuse.FileSystemInterface interace
type Dpfs struct {
	fuse.FileSystemBase
	config            *duplicacy.Config
	storage           duplicacy.Storage
	chunkOperator     *duplicacy.ChunkOperator
	chunkDownloader   *duplicacy.ChunkDownloader
	root              string
	snapshotid        string
	revision          int
	password          string
	preference        *duplicacy.Preference
	repository        string
	mu                sync.Mutex
	cache             DpfsKvStore
	verifiedRevisions map[revisionCacheKey]bool

	// Cache a single snapshot
	lastSnap *duplicacy.Snapshot

	// Cache some data chunks
	chunkCache *lru.Cache[string, *duplicacy.Chunk]

	// Cache backup manager for a snapshot
	lastBackupManager *duplicacy.BackupManager
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
	chunkCache, err := lru.New[string, *duplicacy.Chunk](100)
	if err != nil {
		log.WithError(err).Fatal("unable to create cache")
	}
	self := Dpfs{
		verifiedRevisions: make(map[revisionCacheKey]bool),
		chunkCache:        chunkCache,
	}
	return &self
}
