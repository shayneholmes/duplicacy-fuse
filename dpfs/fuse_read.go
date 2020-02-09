package dpfs

import (
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

// Read satisfies the Read implementation from fuse.FileSystemInterface
func (self *Dpfs) Read(path string, buff []byte, ofst int64, fh uint64) (n int) {
	logger := log.WithFields(
		log.Fields{
			"op":   "Read",
			"path": path,
			"id":   uuid.NewV4().String(),
		})

	info := self.newpathInfo(path)

	files, err := self.getRevisionFiles(info.snapshotid, info.revision)
	if err != nil {
		logger.WithError(err).Debug()
		return 0
	}

	entry, err := self.findFile(info.filepath, files)
	if err != nil {
		logger.WithError(err).Debug()
		return 0
	}

	manager, err := self.createBackupManager(info.snapshotid)
	if err != nil {
		logger.WithError(err).Debug()
		return 0
	}

	snap, err := self.downloadSnapshot(manager, info.snapshotid, info.revision, nil)
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
