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

		// Regex to match current dir and files but not within subdirs
		match := fmt.Sprintf("^%s/[^/]*/?$", info.String())
		regex, err := regexp.Compile(match)
		if err != nil {
			snaplogger.WithError(err).Info()
			return 0
		}

		/* ts := time.Now()
		prefix := key(info.snapshotid, info.revision, info.filepath)
		snaplogger.WithField("prefix", string(prefix)).Info()
		if err := self.cache.Scan(prefix, func(key []byte) error {
			filepath := strings.Join(strings.Split(string(key), ":")[2:], "")
			thisPath := self.abs(filepath, info.snapshotid, info.revision)
			if regex.MatchString(thisPath) {
				_ = strings.TrimPrefix(strings.TrimSuffix(thisPath, "/"), info.String()+"/")
				// fill(pathname, nil, 0)
			}
			return nil
		}); err != nil {
			snaplogger.WithError(err).Info()
		}
		snaplogger.WithField("time", time.Now().Sub(ts)).Info("kv scan time") */

		// do old version of cache for rest of readdir
		files, err := self.getRevisionFiles(info.snapshotid, info.revision)
		if err != nil {
			snaplogger.WithError(err).Debug()
			return 0
		}
		snaplogger.Debug("getRevisionFiles done")

		//ts = time.Now()
		for _, v := range files {
			thisPath := self.abs(v.Path, info.snapshotid, info.revision)
			snaplogger = snaplogger.WithFields(log.Fields{
				"v.Path":   v.Path,
				"thisPath": thisPath,
				"match":    match,
			})
			// snaplogger.Debug()
			if regex.MatchString(thisPath) {
				// Found a match so do fill of dir entry
				pathname := strings.TrimPrefix(strings.TrimSuffix(thisPath, "/"), info.String()+"/")
				snaplogger.WithField("pathname", pathname).Debug("matched")
				fill(pathname, nil, 0)
			}
		}
		//snaplogger.WithField("time", time.Now().Sub(ts)).Info("slice scan time")

		return 0
	}

	// no a revision so list dirs (snapshot-id or revision list)
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
