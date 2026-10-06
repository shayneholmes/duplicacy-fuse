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

func parseRevision(component string) (rev int, err error) {
	rev, err = strconv.Atoi(component)
	if err != nil {
		return
	}
	if rev < 1 {
		err = fmt.Errorf("%s is not a valid revision number", component)
		return
	}
	return
}

// newpathInfo takes a filepath and derives the snapshotid, revision and path taking into account
// the "root" of the mount in self.snapshotid and self.revision. Returns an
// error if the revision is invalid — that is, path provides it, but it is
func (self *Dpfs) newpathInfo(filepath string) (p pathInfo, err error) {
	split := strings.Split(strings.TrimPrefix(filepath, "/"), "/")

	if self.snapshotid != "" {
		p.snapshotid = self.snapshotid
	} else if len(split) > 0 {
		p.snapshotid, split = split[0], split[1:]
	}

	if self.revision != 0 {
		p.revision = self.revision
	} else if len(split) > 0 {
		p.revision, err = parseRevision(split[0]) // This will set an error if it's an invalid value
		split = split[1:]
	} else {
		// Revision isn't specified; leave it as 0
		return
	}

	p.filepath = strings.Join(split, "/")

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

func (self *Dpfs) cacheSnapshotRevisions(snapshotid string) error {
	cacheKey := []byte(fmt.Sprintf("snapshot-revisions:%s", snapshotid))
	logger := log.WithFields(log.Fields{
		"snapshotid": snapshotid,
		"cacheKey":   string(cacheKey),
		"func":       "cacheSnapshotRevisions",
	})
	logger.Debug("caching revisions")
	if self.cache == nil {
		return fmt.Errorf("cache was nil")
	}
	if self.cache.Has(cacheKey) {
		return nil
	}

	// Snapshot revisions aren't in cache: Fetch them.
	manager, err := self.createBackupManager()
	if err != nil {
		return fmt.Errorf("problem creating manager: %w", err)
	}
	revs, err := manager.SnapshotManager.ListSnapshotRevisions(snapshotid)
	if err != nil {
		return fmt.Errorf("problem listing snapshot revisions: %w", err)
	}
	for _, rev := range revs {
		err := self.cacheRevisionInfo(manager, snapshotid, rev)
		if err != nil {
			return fmt.Errorf("problem caching snapshot %v, revision %v: %w", snapshotid, rev, err)
		}
	}

	self.mu.Lock()
	defer self.mu.Unlock()
	if err := self.cache.PutString(cacheKey, "sentinel"); err != nil {
		return fmt.Errorf("problem with Put(%s): %w", cacheKey, err)
	}
	return nil
}

func (self *Dpfs) cacheRevisionInfo(manager *duplicacy.BackupManager, snapshotid string, revision int) error {
	cacheKey := []byte(fmt.Sprintf("revision-info:%s:%d", snapshotid, revision))
	logger := log.WithFields(log.Fields{
		"snapshotid": snapshotid,
		"revision":   revision,
		"cacheKey":   string(cacheKey),
		"func":       "cacheRevisionInfo",
	})
	if self.cache == nil {
		return fmt.Errorf("cache was nil")
	}

	if self.cache.Has(cacheKey) {
		return nil
	}

	// The revision info isn't in the cache: Fetch it and load it
	self.mu.Lock()
	defer self.mu.Unlock()

	if self.cache.Has(cacheKey) {
		logger.Debug("already cached")
		return nil
	}
	logger.Debug("not cached")

	// Retrieve files
	snap, err := self.downloadSnapshotInfo(manager, snapshotid, revision, nil, true)
	if err != nil {
		return fmt.Errorf("problem dowloading snapshot: %w", err)
	}

	if err := self.cache.PutSnapshot(cacheKey, snap); err != nil {
		return fmt.Errorf("problem with Put(%s): %w", cacheKey, err)
	}

	return nil
}

func (self *Dpfs) cacheRevisionFiles(snapshotid string, revision int) error {
	self.mu.Lock()
	defer self.mu.Unlock()

	revisionCacheKey := revisionCacheKey{
		snapshotid: snapshotid,
		revision:   revision,
	}
	if self.verifiedRevisions[revisionCacheKey] {
		return nil
	}
	logger := log.WithFields(log.Fields{
		"snapshotid": snapshotid,
		"revision":   revision,
		"func":       "cacheRevisionFiles",
	})
	is_cached_key := []byte(fmt.Sprintf("%s:%d-iscached", snapshotid, revision))
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
	manager, err := self.createBackupManager()
	if err != nil {
		return fmt.Errorf("problem creating manager: %w", err)
	}
	snap, err := self.downloadSnapshot(manager, snapshotid, revision, nil, true)
	if err != nil {
		return fmt.Errorf("problem dowloading snapshot: %w", err)
	}

	logger.
		WithField("snap.NumberOfFiles", snap.NumberOfFiles).
		WithField("sequencelength", len(snap.FileSequence)).
		Debug("caching files")
	maxSize := 0
	fileCount := 0
	dirCount := 0
	entryCount := 0

	batch := self.cache.CreateEntryBatch()

	entriesByPath := make(map[string][]*duplicacy.Entry)

	snap.ListRemoteFiles(self.config, self.chunkOperator, func(entry *duplicacy.Entry) bool {
		entryCount++
		trimmedPath := strings.Trim(entry.Path, "/")
		dir, _ := path.Split(trimmedPath)
		dir = strings.TrimSuffix(dir, "/")
		if entry.IsDir() {
			dirCount++
			entriesByPath[trimmedPath] = entriesByPath[trimmedPath]
		} else {
			fileCount++
		}
		entriesByPath[dir] = append(entriesByPath[dir], entry)
		return true
	})

	logger.
		WithField("fileCount", fileCount).
		WithField("dirCount", dirCount).
		WithField("entryCount", entryCount).
		WithField("pathCount", len(entriesByPath)).
		WithField("rootPathSize", len(entriesByPath[""])).
		Debug("tabulated all files")

	for path, entries := range entriesByPath {
		k := key(snapshotid, revision, path)
		if n, err := batch.PutEntries(k, entries); err != nil {
			log.
				WithField("key", k).
				WithField("size", n).
				WithError(err).
				Debug("Error inserting key")
			break
		} else {
			if n > maxSize {
				maxSize = n
			}
		}
	}

	self.cache.WriteEntriesBatch(batch)

	logger.
		WithField("largestEntry", maxSize).
		Debug("done caching files")

	if err := self.cache.PutString(is_cached_key, isCached); err != nil {
		return fmt.Errorf("problem with Put(%s): %w", is_cached_key, err)
	}
	self.verifiedRevisions[revisionCacheKey] = true

	return nil
}

func (self *Dpfs) createBackupManager() (*duplicacy.BackupManager, error) {
	// Snapshot IDs are used only during backup or restore, and we're doing
	// neither. As long as the storage stays constant, we want to reuse the
	// backup manager. Additionally, creating a new backup manager for the same
	// storage reconfigures the nesting levels, which isn't thread-safe: it
	// temporarily resets the nesting levels to `nil`, which causes problems if
	// another operation is going on.
	EMPTY_SNAPSHOT_ID := ""

	if self.lastBackupManager != nil {
		return self.lastBackupManager, nil
	}
	manager := duplicacy.CreateBackupManager(EMPTY_SNAPSHOT_ID, self.storage, self.repository, self.password, self.preference.NobackupFile, self.preference.FiltersFile, false)
	if manager == nil {
		return nil, fmt.Errorf("manager was nil")
	}
	if !manager.SetupSnapshotCache(self.preference.Name) {
		return nil, fmt.Errorf("SetupSnapshotCache was false")
	}

	self.lastBackupManager = manager

	return manager, nil
}

func (self *Dpfs) downloadSnapshotInfo(manager *duplicacy.BackupManager, snapshotid string, revision int, patterns []string, attributesNeeded bool) (*duplicacy.Snapshot, error) {
	k := revisionCacheKey{snapshotid, revision}
	if snap, ok := self.snapshotCache.Get(k); ok {
		return snap, nil
	}
	snap := manager.SnapshotManager.DownloadSnapshot(snapshotid, revision)
	if snap == nil {
		return nil, fmt.Errorf("snap was nil")
	}
	// Don't populate the cache until the full info is present.
	return snap, nil
}

func (self *Dpfs) downloadSnapshot(manager *duplicacy.BackupManager, snapshotid string, revision int, patterns []string, attributesNeeded bool) (*duplicacy.Snapshot, error) {
	k := revisionCacheKey{snapshotid, revision}
	if snap, ok := self.snapshotCache.Get(k); ok {
		return snap, nil
	}
	snap, err := self.downloadSnapshotInfo(manager, snapshotid, revision, patterns, attributesNeeded)
	if err != nil {
		return nil, err
	}
	if !manager.SnapshotManager.DownloadSnapshotSequences(snap) {
		return nil, fmt.Errorf("DownloadSnapshotSequences was false")
	}

	self.snapshotCache.Add(k, snap)

	return snap, nil
}

func (self *Dpfs) findFile(snapshotid string, revision int, filepath string) (*duplicacy.Entry, error) {
	// should we update our cache here?
	// this should never be run before something that caches revision contents

	// Use our cache
	dir, _ := path.Split(strings.Trim(filepath, "/"))
	key := key(snapshotid, revision, strings.TrimSuffix(dir, "/"))
	if entries, err := self.cache.GetEntries(key); err != nil {
		return &duplicacy.Entry{}, err
	} else {
		// find the matching entry
		for _, entry := range entries {
			candidatePath := strings.TrimSuffix(entry.Path, "/")
			if candidatePath == filepath {
				return entry, nil
			}
		}
	}
	return &duplicacy.Entry{}, fmt.Errorf("file not found in this path")
}
