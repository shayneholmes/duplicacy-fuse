package main

import (
	"strings"

	"github.com/billziss-gh/cgofuse/fuse"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

func (self *Dpfs) Readdir(path string,
	fill func(name string, stat *fuse.Stat_t, ofst int64) bool,
	ofst int64,
	fh uint64) (errc int) {

	// current and previous
	fill(".", nil, 0)
	fill("..", nil, 0)

	self.lock.RLock()
	defer self.lock.RUnlock()

	id := uuid.NewV4().String()
	sp := self.snapshotPath(path)

	logger := log.WithField("path", path).WithField("sp", sp).WithField("op", "Readdir").WithField("uuid", id)

	snapshotid, revision, _, err := self.info(sp)
	if err != nil {
		logger.WithError(err).WithField("errc", -fuse.ENOSYS).Warning("error listing files")
		return -fuse.ENOSYS
	}

	logger.Debug("start")

	// are we loading from a revision
	if revision != 0 {
		snaplogger := logger.WithField("snapshotid", snapshotid).WithField("revision", revision)
		snaplogger.WithField("call", "CreateSnapshotManager").Debug()

		files, err := self.getRevisionFiles(snapshotid, revision, snaplogger)
		if err != nil {
			snaplogger.WithError(err).Debug()
			return 0
		}

		snaplogger.Debug("loop")
		for _, v := range files {
			slashes := strings.Count(v.Path, "/")
			if slashes > 1 {
				continue
			}

			if slashes == 1 && !strings.HasSuffix(v.Path, "/") {
				continue
			}

			snaplogger.Debug(v.Path)
			fill(strings.TrimSuffix(v.Path, "/"), nil, 0)
		}
		snaplogger.Debug("done")
		return 0
	}

	files, _, err := self.storage.ListFiles(0, sp+"/")
	if err != nil {
		logger.WithError(err).WithField("errc", -fuse.ENOSYS).Warning("error listing files")
		return -fuse.ENOSYS
	}

	logger.WithField("filecount", len(files)).Debug("loop")

	for _, v := range files {
		fill(strings.TrimSuffix(v, "/"), nil, 0)
	}

	logger.WithField("errc", 0).Debug("finish")

	return 0
}
