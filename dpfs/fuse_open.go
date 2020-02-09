package dpfs

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

	info := self.newpathInfo(path)

	files, err := self.getRevisionFiles(info.snapshotid, info.revision)
	if err != nil {
		logger.WithError(err).Debug()
		return -fuse.ENOSYS, ^uint64(0)
	}

	entry, err := self.findFile(info.filepath, files)
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
