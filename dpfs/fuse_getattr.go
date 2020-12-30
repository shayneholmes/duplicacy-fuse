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

	// handle root and first level
	if info.filepath == "" {
		exists, _, _, err := self.storage.GetFileInfo(0, info.String())
		if !exists || err != nil {
			logger.WithError(err).Warning("not a valid snapshot or revision")
			return NoSuchFileOrDirectory
		}
		logger.Debug("is root or first level")
		stat.Mode = fuse.S_IFDIR | 0555
		return 0
	}

	entry, err := self.findFile(info.snapshotid, info.revision, info.filepath)
	if err != nil {
		logger.WithError(err).Debug()
		return NoSuchFileOrDirectory
	}

	if entry.IsDir() {
		logger.Debug("directory")
		stat.Mode = fuse.S_IFDIR | entry.Mode
	} else {
		logger.WithField("size", entry.Size).Debug("file")
		stat.Mode = fuse.S_IFREG | entry.Mode
		stat.Size = entry.Size
	}

	stat.Mtim = fuse.Timespec{
		Sec: entry.Time,
	}

	return 0
}
