package main

import (
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

func (self *Dpfs) Read(path string, buff []byte, ofst int64, fh uint64) (n int) {
	log.SetLevel(log.DebugLevel)
	defer log.SetLevel(log.InfoLevel)

	logger := log.WithFields(
		log.Fields{
			"op":   "Read",
			"path": path,
			"id":   uuid.NewV4().String(),
		})

	snapshotid, revision, p, err := self.info(path)
	if err != nil {
		logger.WithError(err).Debug()
		return 0
	}

	files, err := self.getRevisionFiles(snapshotid, revision)
	if err != nil {
		logger.WithError(err).Debug()
		return 0
	}

	entry, err := self.findFile(p, files)
	if err != nil {
		logger.WithError(err).Debug()
		return 0
	}

	manager, err := self.createBackupManager(snapshotid)
	if err != nil {
		logger.WithError(err).Debug()
		return 0
	}

	snap, err := self.downloadSnapshot(manager, snapshotid, revision, nil)
	if err != nil {
		logger.WithError(err).Debug()
		return 0
	}

	if !manager.SnapshotManager.RetrieveFile(snap, entry, func(chunck []byte) {
		endofst := ofst + int64(len(buff))
		if endofst > int64(len(chunck)) {
			endofst = int64(len(chunck))
		}
		if endofst < ofst {
			n = 0
			return
		}
		n = copy(buff, chunck[ofst:endofst])
	}) {
		return 0
	}

	logger.WithField("n", n).Debug()

	return
}
