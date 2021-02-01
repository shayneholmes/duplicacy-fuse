package dpfs

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/billziss-gh/cgofuse/fuse"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

// Readdir satisfies the Readdir implementation from fuse.FileSystemInterface
func (self *Dpfs) Readdir(path string,
	fill func(name string, stat *fuse.Stat_t, ofst int64) bool,
	ofst int64,
	fh uint64) (errc int) {

	// current and previous
	fill(".", nil, 0)
	fill("..", nil, 0)

	info := self.newpathInfo(path)

	logger := log.WithField("path", path).WithField("op", "Readdir").WithField("uuid", uuid.NewV4().String())

	// are we loading from a revision
	if info.revision != 0 {
		snaplogger := logger.WithField("snapshotid", info.snapshotid).WithField("revision", info.revision)
		snaplogger.WithField("call", "CreateSnapshotManager").Debug()

		// update cache if required
		err := self.cacheRevisionFiles(info.snapshotid, info.revision)
		if err != nil {
			snaplogger.WithError(err).Debug("cacheRevisionFiles")
			return 0
		}
		snaplogger.Debug("cacheRevisionFiles done")

		prefix := key(info.snapshotid, info.revision, strings.TrimPrefix(info.filepath, "/"))

		// For non-root paths in a revision do extra checks
		if info.filepath != "" && info.filepath != "/" {
			// Make sure it actually exists
			entry, err := self.findFile(info.snapshotid, info.revision, info.filepath)
			if err != nil {
				return NoSuchFileOrDirectory
			}

			// Make sure its a dir
			if !entry.IsDir() {
				return NotDirectory
			}

			// Since this is a non-root path, we can add a slash to the prefix to
			// find only items inside the path (and not the path itself).
			prefix = append(prefix, '/')
		}

		snaplogger.WithField("prefix", string(prefix)).Debug()
		if err := self.cache.Scan(prefix, func(key []byte) error {
			relativePath := string(key[len(prefix):])
			snaplogger = snaplogger.WithFields(log.Fields{
				"key":          string(key),
				"relativePath": relativePath,
			})
			if strings.ContainsRune(relativePath, '/') {
				snaplogger.Debug("skipping: entry is inside a subdirectory")
			} else {
				snaplogger.Debug("matched")
				fill(relativePath, nil, 0)
			}
			return nil
		}); err != nil {
			snaplogger.WithError(err).Warning()
		}

		return 0
	}

	// List revisions in a snapshot
	if info.snapshotid != "" {
		// First, make sure that this snapshot has its revisions cached.
		if err := self.cacheSnapshotRevisions(info.snapshotid); err != nil {
			logger.WithError(err).Warning("error caching snapshot revisions")
			return NotImplemented
		}
		prefix := []byte(fmt.Sprintf("revision-info:%s:", info.snapshotid))
		if err := self.cache.Scan(prefix, func(key []byte) error {
			snap, err := self.cache.GetSnapshot(key)
			if err != nil {
				return err
			}
			fill(strconv.FormatInt(int64(snap.Revision), 10), nil, 0)
			logger.WithField("revision", snap.Revision).Debug("matched")
			return nil
		}); err != nil {
			logger.WithError(err).Warning()
		}
		return 0
	}

	// List snapshots directly from the storage
	files, _, err := self.storage.ListFiles(0, info.String()+"/")
	if err != nil {
		logger.WithError(err).Warning("error listing files")
		return NotImplemented
	}

	for _, v := range files {
		logger.WithField("file", v).Debug()
		fill(strings.TrimSuffix(v, "/"), nil, 0)
	}

	return 0
}
