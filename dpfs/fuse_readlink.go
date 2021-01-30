package dpfs

import (
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

// Readlink satisfies the Readlink implementation from fuse.FileSystemInterface
func (self *Dpfs) Readlink(path string) (errc int, link string) {
	errc = NoSuchFileOrDirectory

	info := self.newpathInfo(path)
	logger := log.WithFields(log.Fields{
		"path":       path,
		"op":         "Readlink",
		"uuid":       uuid.NewV4().String(),
		"snapshotid": info.snapshotid,
		"revision":   info.revision,
		"filepath":   info.filepath,
		"id":         uuid.NewV4().String(),
	})

	if info.filepath == "" || info.filepath == "/" {
		// root and first level can't be symlinks
		return
	}

	entry, err := self.findFile(info.snapshotid, info.revision, info.filepath)
	if err != nil {
		logger.WithError(err).Debug()
		return
	}

	if !entry.IsLink() {
		return
	}

	return 0, entry.Link
}
