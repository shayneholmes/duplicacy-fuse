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

func (self *Dpfs) getRevisionFiles(snapshotid string, revision int) ([]*duplicacy.Entry, error) {
	self.lock.RLock()
	defer self.lock.RUnlock()

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

	// Retrieve files
	manager, err := self.createBackupManager(snapshotid)
	if err != nil {
		return nil, fmt.Errorf("problem creating manager: %w", err)
	}
	snap, err := self.downloadSnapshot(manager, snapshotid, revision, nil)
	if err != nil {
		return nil, fmt.Errorf("problem dowloading snapshot: %w", err)
	}

	// Store for later
	self.files.Store(key, readdirCache{
		files: snap.Files,
		ts:    time.Now(),
	})

	return snap.Files, nil
}

func (self *Dpfs) getRevisionFile(snapshotid string, revision int, file string, logger *log.Entry) (buf []byte, err error) {
	/* self.lock.RLock()
	defer self.lock.RUnlock()

	if logger == nil {
		logger = log.WithFields(
			log.Fields{
				"snapshotid": snapshotid,
				"revision":   revision,
			})
	} */
	return
}

func (self *Dpfs) snapshotPath(p string) string {
	self.lock.RLock()
	defer self.lock.RUnlock()

	return strings.TrimSuffix(path.Join(self.root, p), "/")
}

func (self *Dpfs) revision(p string) (revision string) {
	self.lock.RLock()
	defer self.lock.RUnlock()

	slice := strings.Split(p, "/")

	if len(slice) < 2 {
		return ""
	}

	if self.root == "snapshots" {
		if len(slice) < 3 {
			return ""
		} else {
			return slice[2]
		}
	}

	return slice[1]
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

func (self *Dpfs) createBackupManager(snapshotid string) (*duplicacy.BackupManager, error) {
	manager := duplicacy.CreateBackupManager(snapshotid, self.storage, self.repository, self.password, self.preference.NobackupFile, self.preference.FiltersFile)
	if manager == nil {
		return nil, fmt.Errorf("manager was nil")
	}
	if !manager.SetupSnapshotCache(self.preference.Name) {
		return nil, fmt.Errorf("SetupSnapshotCache was false")
	}

	return manager, nil
}

func (self *Dpfs) downloadSnapshot(manager *duplicacy.BackupManager, snapshotid string, revision int, patterns []string) (*duplicacy.Snapshot, error) {
	// Lock so only one DownloadSnapshot is acting at once
	self.flock.Lock()
	defer self.flock.Unlock()

	snap := manager.SnapshotManager.DownloadSnapshot(snapshotid, revision)
	if snap == nil {
		return nil, fmt.Errorf("snap was nil")
	}
	if !manager.SnapshotManager.DownloadSnapshotContents(snap, patterns, true) {
		return nil, fmt.Errorf("DownloadSnapshotContents was false")
	}

	return snap, nil
}

func (self *Dpfs) findFile(filepath string, files []*duplicacy.Entry) (*duplicacy.Entry, error) {
	for _, entry := range files {
		if filepath == entry.Path || filepath+"/" == entry.Path {
			return entry, nil
		}
	}
	return nil, fmt.Errorf("file not found")
}
