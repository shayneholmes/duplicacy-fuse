package main

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"

	"github.com/billziss-gh/cgofuse/fuse"
	duplicacy "github.com/gilbertchen/duplicacy/src"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

const (
	contents = "hello, world\n"
)

// Dpfs is the Duplicacy filesystem type
type Dpfs struct {
	fuse.FileSystemBase
	config  *duplicacy.Config
	storage duplicacy.Storage
	lock    sync.RWMutex
	root    string
}

func NewDuplicacyfs() *Dpfs {
	self := Dpfs{}
	return &self
}

func (self *Dpfs) Init() {
	var repository, storageName, storagePassword string
	var snapshot int
	var all bool

	self.lock.Lock()
	defer self.lock.Unlock()

	_, err := fuse.OptParse(os.Args, "repository=%s storage=%s snapshot=%d password=%s all", &repository, &storageName, &snapshot, &storagePassword, &all)
	if err != nil {
		log.WithError(err).Fatal("arg error")
	}

	// Set defaults if unspecified
	if repository == "" {
		repository = "."
	}

	if storageName == "" {
		storageName = "default"
	}

	if !duplicacy.LoadPreferences(repository) {
		log.WithField("repository", repository).Fatal("problem loading preferences")
	}

	preferencePath := duplicacy.GetDuplicacyPreferencePath()
	duplicacy.SetKeyringFile(path.Join(preferencePath, "keyring"))

	preference := duplicacy.FindPreference(storageName)
	if preference == nil {
		log.WithField("storageName", storageName).Fatal("storage not found")
	}

	self.storage = duplicacy.CreateStorage(*preference, false, 1)
	if self.storage == nil {
		log.WithField("storageName", storageName).Fatal("could not create storage")
	}

	if preference.Encrypted && storagePassword == "" {
		log.WithField("storageName", storageName).Fatal("storage is encrypted but no password provided")
	}

	if all {
		self.root = "snapshots"
	} else {
		self.root = path.Join("snapshots", preference.SnapshotID)
	}

	config, _, err := duplicacy.DownloadConfig(self.storage, storagePassword)
	if err != nil {
		log.WithField("storageName", storageName).WithError(err).Fatal("failed to download config from storage")
	}

	if config == nil {
		log.WithField("storageName", storageName).Fatal("storage is not configured")
	}

	self.config = config
}

func (self *Dpfs) Open(path string, flags int) (errc int, fh uint64) {
	switch path {
	case "/hello":
		return 0, 0
	default:
		return -fuse.ENOENT, ^uint64(0)
	}
}

func (self *Dpfs) Getattr(path string, stat *fuse.Stat_t, fh uint64) (errc int) {
	sp := self.snapshotPath(path)
	id := uuid.NewV4().String()
	logger := log.WithField("path", path).WithField("sp", sp).WithField("op", "Getattr").WithField("uuid", id)
	logger.Debug()

	// handle root and first level
	if strings.Count(sp, "/") <= 2 {
		logger.WithField("errc", 0).Debug("is root or first level")
		stat.Mode = fuse.S_IFDIR | 0555
		return 0
	}

	self.lock.RLock()
	defer self.lock.RUnlock()

	// get file info
	exists, isDir, size, err := self.storage.GetFileInfo(0, sp)
	if err != nil {
		logger.WithError(err).WithField("errc", -fuse.ENOSYS).Debug()
		return -fuse.ENOSYS
	}

	if !exists {
		logger.WithField("errc", -fuse.ENOENT).Debug("does not exist")
		return -fuse.ENOENT
	}

	if isDir {
		logger.Debug("directory")
		stat.Mode = fuse.S_IFDIR | 0555
	} else {
		logger.WithField("size", size).Debug("file")
		stat.Mode = fuse.S_IFREG | 0444
		stat.Size = size
	}

	logger.WithField("errc", 0).Debug("end")
	return 0
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
	if strings.HasPrefix(p, self.root) {
		return p
	}

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

func (self *Dpfs) info(p string) (snapshotid string, revision int, err error) {
	self.lock.RLock()
	defer self.lock.RUnlock()

	switch v := strings.Split(self.snapshotPath(p), "/"); len(v) {
	case 0:
		err = fmt.Errorf("invalid path")
	case 1:
		snapshotid = ""
		revision = 0
	case 2:
		snapshotid = v[1]
		revision = 0
	default:
		snapshotid = v[1]
		revision, err = strconv.Atoi(v[2])
	}

	return
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
