package main

import (
	"strings"

	"github.com/billziss-gh/cgofuse/fuse"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

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
