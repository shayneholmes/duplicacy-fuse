package main

import (
	"time"

	"github.com/billziss-gh/cgofuse/fuse"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

func (self *Dpfs) Getattr(path string, stat *fuse.Stat_t, fh uint64) (errc int) {
	self.lock.RLock()
	defer self.lock.RUnlock()

	if path == "/desktop.ini" ||
		path == "/folder.jpg" ||
		path == "/folder.gif" {
		return -fuse.ENOSYS
	}

	sp := self.snapshotPath(path)
	id := uuid.NewV4().String()
	logger := log.WithField("path", path).WithField("sp", sp).WithField("op", "Getattr").WithField("uuid", id)

	snapshotid, revision, p, err := self.info(path)
	if err != nil {
		logger.WithError(err).Debug()
		return -fuse.ENOSYS
	}
	logger = logger.WithFields(log.Fields{
		"snapshotid": snapshotid,
		"revision":   revision,
		"p":          p,
	})

	// handle root and first level
	if p == "" {
		logger.Debug("is root or first level")
		stat.Mode = fuse.S_IFDIR | 0555
		return 0
	}

	files, err := self.getRevisionFiles(snapshotid, revision)
	if err != nil {
		logger.WithError(err).Debug()
		return -fuse.ENOSYS
	}

	startts := time.Now()
	entry, err := self.FindFile(p, files)
	logger.WithField("loop time", time.Since(startts).String()).Debug()
	if err != nil {
		return -fuse.ENOENT
	}

	if entry.IsDir() {
		logger.Debug("directory")
		stat.Mode = fuse.S_IFDIR | 0555
	} else {
		logger.WithField("size", entry.Size).Debug("file")
		stat.Mode = fuse.S_IFREG | 0444
		stat.Size = entry.Size
	}

	return 0
}
