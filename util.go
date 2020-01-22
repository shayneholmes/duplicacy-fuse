package main

import (
	"fmt"

	duplicacy "github.com/gilbertchen/duplicacy/src"
	log "github.com/sirupsen/logrus"
)

func (self *Dpfs) getRevisionFiles(snapshotid string, revision int, logger *log.Entry) ([]*duplicacy.Entry, error) {
	self.lock.RLock()
	defer self.lock.RUnlock()

	if logger == nil {
		logger = log.WithField("snapshotid", snapshotid).WithField("revision", revision)
	}

	// Check for cached list of files
	key := fmt.Sprintf("%s_%d", snapshotid, revision)
	if f, ok := self.files.Load(key); ok {
		return f.([]*duplicacy.Entry), nil
	}

	// Retrieve files
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

	// Store for later
	self.files.Store(key, snap.Files)

	return snap.Files, nil
}
