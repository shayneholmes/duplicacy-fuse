package main

import (
	"os"
	"path"
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
	manager *duplicacy.BackupManager
	storage duplicacy.Storage
	lock    sync.RWMutex
}

func NewDuplicacyfs() *Dpfs {
	self := Dpfs{}
	return &self
}

func (self *Dpfs) Init() {
	var repository, storageName string
	var snapshot int

	self.lock.Lock()
	defer self.lock.Unlock()

	_, err := fuse.OptParse(os.Args, "repository=%s storage=%s snapshot=%d", &repository, &storageName, &snapshot)
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

	self.storage = duplicacy.CreateStorage(*preference, false, 4)
	if self.storage == nil {
		log.WithField("storageName", storageName).Fatal("could not create storage")
	}

	self.manager = duplicacy.CreateBackupManager(preference.SnapshotID, self.storage, repository, "", "", "")

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
	sp := snapshotPath(path)
	id := uuid.NewV4().String()
	logger := log.WithField("path", path).WithField("sp", sp).WithField("op", "Getattr").WithField("uuid", id)
	logger.Info()

	// handle root and first level
	if strings.Count(sp, "/") <= 2 {
		logger.WithField("errc", 0).Info("is root or first level")
		stat.Mode = fuse.S_IFDIR | 0555
		return 0
	}

	self.lock.RLock()
	defer self.lock.RUnlock()

	// get file info
	exists, isDir, size, err := self.storage.GetFileInfo(0, sp)
	if err != nil {
		logger.WithError(err).WithField("errc", -fuse.ENOSYS).Info()
		return -fuse.ENOSYS
	}

	if !exists {
		logger.WithField("errc", -fuse.ENOENT).Info("does not exist")
		return -fuse.ENOENT
	}

	if isDir {
		logger.Info("directory")
		stat.Mode = fuse.S_IFDIR | 0555
	} else {
		logger.WithField("size", size).Info("file")
		stat.Mode = fuse.S_IFREG | 0444
		stat.Size = size
	}

	logger.WithField("errc", 0).Info("end")
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

func (self *Dpfs) Readdir(path string,
	fill func(name string, stat *fuse.Stat_t, ofst int64) bool,
	ofst int64,
	fh uint64) (errc int) {
	fill(".", nil, 0)
	fill("..", nil, 0)
	self.lock.RLock()
	defer self.lock.RUnlock()
	sp := snapshotPath(path)
	id := uuid.NewV4().String()
	logger := log.WithField("path", path).WithField("sp", sp).WithField("op", "Readdir").WithField("uuid", id)

	logger.Info("start")

	// This is a snapshot path
	if strings.Count(sp, "/") == 2 {
		manager := duplicacy.CreateSnapshotManager(self.config)
	}

	files, _, err := self.storage.ListFiles(0, sp)
	if err != nil {
		logger.WithError(err).WithField("errc", -fuse.ENOSYS).Info("error listing files")
		return -fuse.ENOSYS
	}

	logger.WithField("filecount", len(files)).Info("loop")

	for _, v := range files {
		fill(strings.TrimSuffix(v, "/"), nil, 0)
	}

	logger.WithField("errc", 0).Info("finish")

	return 0
}

func snapshotPath(p string) string {
	return strings.TrimSuffix(path.Join("snapshots", p), "/")
}

func main() {
	if len(os.Args) <= 1 {
		log.Fatal("missing mountpoint")
	}

	duplicacyfs := NewDuplicacyfs()
	host := fuse.NewFileSystemHost(duplicacyfs)
	host.Mount("", os.Args[1:])
}
