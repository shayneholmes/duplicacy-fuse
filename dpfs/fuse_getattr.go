package dpfs

import (
	"github.com/billziss-gh/cgofuse/fuse"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

// Getattr satisfies the Getattr implementation from fuse.FileSystemInterface
func (self *Dpfs) Getattr(path string, stat *fuse.Stat_t, fh uint64) (errc int) {
	info := self.newpathInfo(path)
	logger := log.WithFields(log.Fields{
		"path":       path,
		"op":         "Getattr",
		"uuid":       uuid.NewV4().String(),
		"snapshotid": info.snapshotid,
		"revision":   info.revision,
		"filepath":   info.filepath,
		"id":         uuid.NewV4().String(),
	})

	// handle files that shouldn't exist here
	if (info.snapshotid == "" && info.revision == 0) &&
		(path == "/desktop.ini" ||
			path == "/folder.jpg" ||
			path == "/folder.gif") {
		return -fuse.ENOENT
	}

	if (info.revision == 0) &&
		(path == "/"+info.snapshotid+"/desktop.ini" ||
			path == "/"+info.snapshotid+"/folder.jpg" ||
			path == "/"+info.snapshotid+"/folder.gif") {
		return -fuse.ENOENT
	}

	// handle root and first level
	if info.filepath == "" {
		logger.Debug("is root or first level")
		stat.Mode = fuse.S_IFDIR | 0555
		return 0
	}

	entry, err := self.findFile(info.snapshotid, info.revision, info.filepath)
	if err != nil {
		logger.WithError(err).Debug()
		return -fuse.ENOENT
	}

	if entry.IsDir() {
		logger.Debug("directory")
		stat.Mode = fuse.S_IFDIR | 0555
	} else {
		logger.WithField("size", entry.Size).Debug("file")
		stat.Mode = fuse.S_IFREG | 0444
		stat.Size = entry.Size
	}

	stat.Mtim = fuse.Timespec{
		Sec: entry.Time,
	}

	/* entry, err := self.findFile(info.filepath, files)
	if err != nil {
		logger.WithError(err).Debug("findFile")
		return -fuse.ENOENT
	}
	if entry.IsDir() {
		logger.Debug("directory")
		stat.Mode = fuse.S_IFDIR | 0555
	} else {
		logger.WithField("size", entry.Size).Debug("file")
		stat.Mode = fuse.S_IFREG | 0444
		stat.Size = entry.Size
	} */

	return 0
}
