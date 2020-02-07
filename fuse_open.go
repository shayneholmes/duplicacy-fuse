package main

import (
	"fmt"

	"github.com/billziss-gh/cgofuse/fuse"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

func (self *Dpfs) Open(path string, flags int) (errc int, fh uint64) {
	logger := log.WithFields(
		log.Fields{
			"op":   "Open",
			"path": path,
			"id":   uuid.NewV4().String(),
		})

	snapshotid, revision, p, err := self.info(path)
	if err != nil {
		logger.WithError(err).Debug()
		return -fuse.ENOSYS, ^uint64(0)
	}

	files, err := self.getRevisionFiles(snapshotid, revision)
	if err != nil {
		logger.WithError(err).Debug()
		return -fuse.ENOSYS, ^uint64(0)
	}

	entry, err := self.findFile(p, files)
	if err != nil {
		logger.WithError(err).Debug()
		return -fuse.ENOENT, ^uint64(0)
	}

	if entry.IsDir() {
		logger.WithError(fmt.Errorf("is a direcotry")).Debug()
		return -fuse.EISDIR, ^uint64(0)
	}

	return 0, 0
}
