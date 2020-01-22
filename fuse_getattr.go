package main

import (
	"github.com/billziss-gh/cgofuse/fuse"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

func (self *Dpfs) Getattr(path string, stat *fuse.Stat_t, fh uint64) (errc int) {
	self.lock.RLock()
	defer self.lock.RUnlock()

	sp := self.snapshotPath(path)
	id := uuid.NewV4().String()
	logger := log.WithField("path", path).WithField("sp", sp).WithField("op", "Getattr").WithField("uuid", id)
	logger.Debug()

	snapshotid, revision, p, err := self.info(path)
	if err != nil {
		logger.WithError(err).WithField("errc", -fuse.ENOSYS).Debug()
		return -fuse.ENOSYS
	}
	logger.WithFields(log.Fields{
		"snapshotid": snapshotid,
		"revision":   revision,
		"p":          p,
	}).Debug()

	// handle root and first level
	if p == "" {
		logger.WithField("errc", 0).Debug("is root or first level")
		stat.Mode = fuse.S_IFDIR | 0555
		return 0
	}

	files, err := self.getRevisionFiles(snapshotid, revision, logger)
	if err != nil {
		logger.WithError(err).WithField("errc", -fuse.ENOSYS).Debug()
		return -fuse.ENOSYS
	}

	for _, v := range files {
		if p == v.Path {
			logger.Debug("found")
			if v.IsDir() {
				logger.Debug("directory")
				stat.Mode = fuse.S_IFDIR | 0555
			} else {
				logger.WithField("size", v.Size).Debug("file")
				stat.Mode = fuse.S_IFREG | 0444
				stat.Size = v.Size
			}
			break
		}
	}

	logger.WithField("errc", 0).Debug("end")
	return 0
}
