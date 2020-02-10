package dpfs

import (
	"fmt"

	"github.com/billziss-gh/cgofuse/fuse"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

// Open satisfies the Open implementation from fuse.FileSystemInterface
func (self *Dpfs) Open(path string, flags int) (errc int, fh uint64) {
	logger := log.WithFields(
		log.Fields{
			"op":   "Open",
			"path": path,
			"id":   uuid.NewV4().String(),
		})

	info := self.newpathInfo(path)

	entry, err := self.findFile(info.snapshotid, info.revision, info.filepath)
	if err != nil {
		logger.WithError(err).Debug()
		return -fuse.ENOENT, 0
	}

	if entry.IsDir() {
		logger.WithError(fmt.Errorf("is a direcotry")).Debug()
		return -fuse.EISDIR, ^uint64(0)
	}

	return 0, 0
}
