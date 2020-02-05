package main

import (
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

func (self *Dpfs) Read(path string, buff []byte, ofst int64, fh uint64) (n int) {
	self.lock.RLock()
	defer self.lock.RUnlock()

	sp := self.snapshotPath(path)
	id := uuid.NewV4().String()
	logger := log.WithField("path", path).WithField("sp", sp).WithField("op", "Open").WithField("uuid", id)

	snapshotid, revision, p, err := self.info(path)
	if err != nil {
		logger.WithError(err).Debug()
		return 0
	}
	logger = logger.WithFields(log.Fields{
		"snapshotid": snapshotid,
		"revision":   revision,
		"p":          p,
	})

	files, err := self.getRevisionFiles(snapshotid, revision)
	if err != nil {
		logger.WithError(err).Debug()
		return 0
	}

	_, err = self.FindFile(path, files)
	if err != nil {
		return 0
	}

	// not sure what to do from here...

	/* endofst := ofst + int64(len(buff))
	if endofst > int64(len(contents)) {
		endofst = int64(len(contents))
	}
	if endofst < ofst {
		return 0
	}
	n = copy(buff, contents[ofst:endofst]) */
	return
}
