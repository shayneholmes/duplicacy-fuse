package dpfs

import (
	"fmt"
	"regexp"
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

		// For non-root paths in a revision do extra checks
		if info.filepath != "" {
			// Make sure it actually exists
			entry, err := self.findFile(info.snapshotid, info.revision, info.filepath)
			if err != nil {
				return NoSuchFileOrDirectory
			}

			// Make sure its a dir
			if entry.IsFile() {
				return NotDirectory
			}
		}

		// Regex to match current dir and files but not within subdirs
		match := fmt.Sprintf("^%s/[^/]*/?$", regexp.QuoteMeta(info.String()))
		regex, err := regexp.Compile(match)
		if err != nil {
			snaplogger.WithError(err).Warning()
			return 0
		}

		prefix := key(info.snapshotid, info.revision, strings.TrimPrefix(info.filepath, "/"))
		snaplogger.WithField("prefix", string(prefix)).Debug()
		if err := self.cache.Scan(prefix, func(key []byte) error {
			filepath := strings.Join(strings.Split(string(key), ":")[2:], "")
			thisPath := self.abs(filepath, info.snapshotid, info.revision)
			snaplogger = snaplogger.WithFields(log.Fields{
				"v.Path":   filepath,
				"thisPath": thisPath,
				"match":    match,
			})
			if regex.MatchString(thisPath) {
				pathname := strings.TrimPrefix(strings.TrimSuffix(thisPath, "/"), info.String()+"/")
				snaplogger.WithField("pathname", pathname).Debug("matched")
				fill(pathname, nil, 0)
			} else {
				snaplogger.Debug("NO match")
			}
			return nil
		}); err != nil {
			snaplogger.WithError(err).Warning()
		}

		return 0
	}

	// not a revision so list dirs (snapshot-id or revision list)
	files, _, err := self.storage.ListFiles(0, info.String()+"/")
	if err != nil {
		logger.WithError(err).Warning("error listing files")
		return -fuse.ENOSYS
	}

	for _, v := range files {
		logger.WithField("file", v).Debug()
		fill(strings.TrimSuffix(v, "/"), nil, 0)
	}

	return 0
}
