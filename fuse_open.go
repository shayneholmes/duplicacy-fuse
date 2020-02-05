package main

import (
	"github.com/billziss-gh/cgofuse/fuse"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

func (self *Dpfs) Open(path string, flags int) (errc int, fh uint64) {
	sp := self.snapshotPath(path)
	id := uuid.NewV4().String()
	logger := log.WithField("path", path).WithField("sp", sp).WithField("op", "Open").WithField("uuid", id)

	snapshotid, revision, p, err := self.info(path)
	if err != nil {
		logger.WithError(err).Debug()
		return -fuse.ENOSYS, ^uint64(0)
	}
	logger = logger.WithFields(log.Fields{
		"snapshotid": snapshotid,
		"revision":   revision,
		"p":          p,
	})

	files, err := self.getRevisionFiles(snapshotid, revision)
	if err != nil {
		logger.WithError(err).Debug()
		return -fuse.ENOSYS, ^uint64(0)
	}

	entry, err := self.FindFile(path, files)
	if err != nil {
		return -fuse.ENOENT, ^uint64(0)
	}

	if entry.IsDir() {
		return -fuse.EISDIR, ^uint64(0)
	}

	// Is any of this needed for a read only fileystem?
	self.lock.Lock()
	self.ino++
	fh = self.ino
	if of, ok := self.ofiles[self.ino]; ok {
		// File is already open so increment opencnt
		of.opencnt++
		self.ofiles[self.ino] = of
	} else {
		// file was not open
		self.ofiles[self.ino] = node_t{
			path:       p,
			revision:   revision,
			snapshotid: snapshotid,
			opencnt:    1,
		}
	}
	self.lock.RUnlock()

	return 0, fh
}
