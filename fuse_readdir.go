package main

import (
	"fmt"
	"strings"

	"github.com/billziss-gh/cgofuse/fuse"
	duplicacy "github.com/gilbertchen/duplicacy/src"
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

func (self *Dpfs) getRevisionFiles(snapshotid string, revision int, logger *log.Entry) ([]*duplicacy.Entry, error) {
	self.lock.RLock()
	defer self.lock.RUnlock()

	if logger == nil {
		logger = log.WithField("snapshotid", snapshotid).WithField("revision", revision)
	}

	logger.WithField("call", "CreateBackupManager").Debug()
	manager := duplicacy.CreateBackupManager(snapshotid, self.storage, self.repository, self.password, self.preference.NobackupFile, self.preference.FiltersFile)
	if manager == nil {
		logger.WithField("call", "CreateBackupManager").Warning("manager was nil")
		return nil, fmt.Errorf("manager was nil")
	}
	logger.WithField("call", "SetupSnapshotCache").Debug()
	manager.SetupSnapshotCache(self.preference.Name)
	logger.WithField("call", "DownloadSnapshot").Debug("before")
	snap := manager.SnapshotManager.DownloadSnapshot(snapshotid, revision)
	logger.WithField("call", "DownloadSnapshot").Debug("after")
	if snap == nil {
		logger.WithField("call", "DownloadSnapshot").Warning("snap was nil")
		return nil, fmt.Errorf("snap was nil")
	}
	logger.WithField("call", "DownloadSnapshotContents").Debug()
	patterns := []string{}
	manager.SnapshotManager.DownloadSnapshotContents(snap, patterns, true)
	if snap == nil {
		logger.WithField("call", "DownloadSnapshotContents").Warning("snap is still nil")
		return nil, fmt.Errorf("snap is still nil")
	}

	return snap.Files, nil
}
