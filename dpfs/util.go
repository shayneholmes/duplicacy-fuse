package dpfs

import (
	"fmt"
	"path"
	"strconv"
	"strings"

	duplicacy "github.com/gilbertchen/duplicacy/src"
	log "github.com/sirupsen/logrus"
)

type pathInfo struct {
	filepath   string
	snapshotid string
	revision   int
}

const isCached = "cached ok"

// newpathInfo takes a filepath and derives the snapshotid, revision and path taking into account
// the "root" of the mount in self.snapshotid and self.revision
func (self *Dpfs) newpathInfo(filepath string) (p pathInfo) {
	// revision and snapshotid is set so filepath is just filepath
	if self.snapshotid != "" && self.revision != 0 {
		p.snapshotid = self.snapshotid
		p.revision = self.revision
		p.filepath = filepath
		return
	}

	// snapshotid is set so filepath may contain revision as first entry followed by filepath
	if self.snapshotid != "" {
		p.snapshotid = self.snapshotid
		if split := strings.Split(filepath, "/"); len(split) > 1 {
			if split[1] != "" {
				p.revision, _ = strconv.Atoi(split[1])
			}
			if len(split) > 2 {
				p.filepath = "/" + strings.Join(split[2:], "/")
			}
		}
		return
	}

	// neither is set so filepath may contain snapshotid followed by revision followed by filepath
	switch split := strings.Split(filepath, "/"); len(split) {
	case 0, 1:
		// this should not happen
		return
	case 2:
		p.snapshotid = split[1]
	case 3:
		p.snapshotid = split[1]
		p.revision, _ = strconv.Atoi(split[2])
	default:
		p.snapshotid = split[1]
		p.revision, _ = strconv.Atoi(split[2])
		p.filepath = "/" + strings.Join(split[3:], "/")
	}

	return
}

// String function for pathInfo
func (info *pathInfo) String() string {
	if info.filepath != "" {
		return fmt.Sprintf("snapshots/%s/%d%s", info.snapshotid, info.revision, info.filepath)
	}
	if info.revision != 0 {
		return fmt.Sprintf("snapshots/%s/%d", info.snapshotid, info.revision)
	}

	if info.snapshotid != "" {
		return fmt.Sprintf("snapshots/%s", info.snapshotid)

	}
	return "snapshots"
}

func (self *Dpfs) cacheRevisionFiles(snapshotid string, revision int) error {
	self.mu.Lock()
	defer self.mu.Unlock()

	logger := log.WithFields(log.Fields{
		"snapshotid": snapshotid,
		"revision":   revision,
		"func":       "cacheRevisionFiles",
	})
	is_cached_key := []byte(fmt.Sprintf("%s:%d", snapshotid, revision))
	if self.cache == nil {
		return fmt.Errorf("cache was nil")
	}
	logger.WithField("is_cached_key", string(is_cached_key)).Debug("checking if cached already")
	if v, err := self.cache.GetString(is_cached_key); err == nil && v == isCached {
		logger.WithField("is_cached_key", string(is_cached_key)).Debug("already cached")
		return nil
	}
	logger.WithField("is_cached_key", string(is_cached_key)).Debug("not cached")

	// Retrieve files
	manager, err := self.createBackupManager(snapshotid)
	if err != nil {
		return fmt.Errorf("problem creating manager: %w", err)
	}
	snap, err := self.downloadSnapshot(manager, snapshotid, revision, nil, true)
	if err != nil {
		return fmt.Errorf("problem dowloading snapshot: %w", err)
	}

	for _, entry := range snap.Files {
		k := key(snapshotid, revision, entry.Path)
		if err != nil {
			return fmt.Errorf("problem encoding entry (%s): %w", k, err)
		}
		if err := self.cache.PutEntry(k, entry); err != nil {
			log.WithError(err).Debug(string(k))
			return fmt.Errorf("problem with Put(%s): %w", k, err)
		}
	}

	if err := self.cache.PutString(is_cached_key, isCached); err != nil {
		return fmt.Errorf("problem with Put(%s): %w", is_cached_key, err)
	}

	return nil
}

// converts to an absolute path
func (self *Dpfs) abs(filepath string, snapshotid string, revision int) (absolutepath string) {
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

func (self *Dpfs) downloadSnapshot(manager *duplicacy.BackupManager, snapshotid string, revision int, patterns []string, attributesNeeded bool) (*duplicacy.Snapshot, error) {
	snap := manager.SnapshotManager.DownloadSnapshot(snapshotid, revision)
	if snap == nil {
		return nil, fmt.Errorf("snap was nil")
	}
	if !manager.SnapshotManager.DownloadSnapshotContents(snap, patterns, attributesNeeded) {
		return nil, fmt.Errorf("DownloadSnapshotContents was false")
	}

	return snap, nil
}

func (self *Dpfs) findFile(snapshotid string, revision int, filepath string) (*duplicacy.Entry, error) {
	// should we update our cache here?
	// this should never be run before something that caches revision contents
	filepath = strings.TrimPrefix(filepath, "/")

	// Use our cache
	return self.cache.GetEntry(key(snapshotid, revision, filepath))
}
