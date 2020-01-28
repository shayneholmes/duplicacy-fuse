package main

import (
	"fmt"
	"path"
	"strconv"
	"strings"
	"time"

	duplicacy "github.com/gilbertchen/duplicacy/src"
	log "github.com/sirupsen/logrus"
)

type readdirCache struct {
	files []*duplicacy.Entry
	ts    time.Time
}

func (self *Dpfs) getRevisionFiles(snapshotid string, revision int, logger *log.Entry) ([]*duplicacy.Entry, error) {
	self.lock.RLock()
	defer self.lock.RUnlock()

	if logger == nil {
		logger = log.WithFields(
			log.Fields{
				"snapshotid": snapshotid,
				"revision":   revision,
			})
	}

	// Check for cached list of files
	key := fmt.Sprintf("%s_%d", snapshotid, revision)
	if f, ok := self.files.Load(key); ok {
		cache, castok := f.(readdirCache)
		if !castok {
			return nil, fmt.Errorf("could not cast result as readdirCache")
		}
		// Update cache timestamp and store
		cache.ts = time.Now()
		self.files.Store(key, cache)

		// return result
		return cache.files, nil
	}

	// Lock so only one BackupManager is acting at once
	self.flock.Lock()
	defer self.flock.Unlock()

	// Retrieve files
	manager := duplicacy.CreateBackupManager(snapshotid, self.storage, self.repository, self.password, self.preference.NobackupFile, self.preference.FiltersFile)
	if manager == nil {
		logger.WithField("call", "CreateBackupManager").Warning("manager was nil")
		return nil, fmt.Errorf("manager was nil")
	}
	manager.SetupSnapshotCache(self.preference.Name)
	snap := manager.SnapshotManager.DownloadSnapshot(snapshotid, revision)
	if snap == nil {
		logger.WithField("call", "DownloadSnapshot").Warning("snap was nil")
		return nil, fmt.Errorf("snap was nil")
	}
	patterns := []string{}
	manager.SnapshotManager.DownloadSnapshotContents(snap, patterns, true)
	if snap == nil {
		logger.WithField("call", "DownloadSnapshotContents").Warning("snap is still nil")
		return nil, fmt.Errorf("snap is still nil")
	}

	// Store for later
	self.files.Store(key, readdirCache{
		files: snap.Files,
		ts:    time.Now(),
	})

	return snap.Files, nil
}

func (self *Dpfs) abs(filepath string, snapshotid string, revision int) (absolutepath string) {
	self.lock.RLock()
	defer self.lock.RUnlock()

	switch strings.Count(self.root, "/") {
	case 0:
		if revision == 0 {
			return path.Join(self.root, snapshotid, filepath)
		}
		return path.Join(self.root, snapshotid, strconv.Itoa(revision), filepath)
	case 1:
		if revision == 0 {
			return path.Join(self.root, filepath)
		}
		return path.Join(self.root, strconv.Itoa(revision), filepath)
	}
	return path.Join(self.root, filepath)
}

func (self *Dpfs) info(p string) (snapshotid string, revision int, path string, err error) {
	self.lock.RLock()
	defer self.lock.RUnlock()

	if !strings.HasPrefix(p, "snapshots") {
		p = self.snapshotPath(p)
	}

	switch v := strings.Split(p, "/"); len(v) {
	case 0:
		err = fmt.Errorf("invalid path")
	case 1:
		snapshotid = ""
		revision = 0
	case 2:
		snapshotid = v[1]
		revision = 0
	default:
		snapshotid = v[1]
		revision, _ = strconv.Atoi(v[2])
		path = strings.Join(v[3:], "/")
	}

	return
}

func (self *Dpfs) cleanReaddirCache(age time.Duration) {

	for {
		// wait for timer to expire
		time.Sleep(time.Second * 30)

		// go through map and remove old items
		self.files.Range(func(key, value interface{}) bool {
			log.WithField("key", key).Debug("checking age of cache")
			if cache, ok := value.(readdirCache); ok {
				purgeafter := time.Now().Add(-age)
				logger := log.WithFields(log.Fields{
					"key":        key,
					"ts":         cache.ts,
					"purgeafter": purgeafter,
				})
				if cache.ts.Before(purgeafter) {
					logger.Debug("purging item")
					self.files.Delete(key)
				} else {
					logger.Debug("keeping item")
				}
			}
			return true
		})
	}
}
