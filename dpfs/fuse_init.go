package dpfs

import (
	"os"
	"path"
	"path/filepath"
	"runtime"
	"time"

	"github.com/billziss-gh/cgofuse/fuse"
	duplicacy "github.com/gilbertchen/duplicacy/src"
	log "github.com/sirupsen/logrus"
)

// Init satisfies the Init implementation from fuse.FileSystemInterface
func (self *Dpfs) Init() {
	var repository, snapshotid, storageName, storagePassword, loglevel, cachedir string
	var revision int
	var debug, all bool

	_, err := fuse.OptParse(os.Args, "repository=%s storage=%s snapshot=%s revision=%d password=%s loglevel=%s cachedir=%s debug all", &repository, &storageName, &snapshotid, &revision, &storagePassword, &loglevel, &cachedir, &debug, &all)
	if err != nil {
		log.WithError(err).Fatal("arg error")
	}

	// enable debug if arg set
	if debug {
		log.SetLevel(log.DebugLevel)
	} else {
		switch loglevel {
		case "debug":
			log.SetLevel(log.DebugLevel)
		case "warn":
			log.SetLevel(log.WarnLevel)
		case "info":
			log.SetLevel(log.InfoLevel)
		}
	}

	// Set defaults if unspecified
	if repository == "" {
		repository, err = os.Getwd()
		if err != nil {
			log.WithError(err).Fatal("could not get current dir")
		}
	}

	// check cachedir and create if required
	if cachedir == "" {
		var homedir string
		if runtime.GOOS == "windows" {
			homedir = os.Getenv("USERPROFILE")
		} else {
			homedir = os.Getenv("HOME")
		}

		cachedir = filepath.Join(homedir, ".duplicacy-fuse")
	}
	if stat, err := os.Stat(cachedir); os.IsNotExist(err) {
		if err := os.Mkdir(cachedir, 0755); err != nil {
			log.WithError(err).Fatal("error creating cache dir")
		}
	} else if !stat.IsDir() {
		log.Fatal("cache dir exists but is not a directory")
	}

	// Create new cache/kv db
	kvpath := "bitcask://" + cachedir
	cache, err := NewDpfsKv(kvpath)
	if err != nil {
		log.WithError(err).Fatal("problem creating kv cache")
	}
	log.WithField("kvpath", kvpath).Info("created kv store")
	self.cache = cache

	if storageName == "" {
		storageName = "default"
	}

	if all && snapshotid != "" {
		log.Fatal("cannot use all and snapshotid at the same time")
	}

	if snapshotid == "" && revision != 0 {
		log.Fatal("cannot specify a revision without a snapshot id")
	}

	if !duplicacy.LoadPreferences(repository) {
		log.WithField("repository", repository).Fatal("problem loading preferences")
	}

	preferencePath := duplicacy.GetDuplicacyPreferencePath()
	duplicacy.SetKeyringFile(path.Join(preferencePath, "keyring"))

	self.preference = duplicacy.FindPreference(storageName)
	if self.preference == nil {
		log.WithField("storageName", storageName).Fatal("storage not found")
	}

	self.storage = duplicacy.CreateStorage(*self.preference, false, 1)
	if self.storage == nil {
		log.WithField("storageName", storageName).Fatal("could not create storage")
	}

	if self.preference.Encrypted && storagePassword == "" {
		log.WithField("storageName", storageName).Fatal("storage is encrypted but no password provided")
	}

	self.root = "snapshots"

	if snapshotid != "" {
		self.snapshotid = self.preference.SnapshotID
	}

	if revision != 0 {
		self.revision = revision
	}

	log.WithField("root", self.root).Debug()

	config, _, err := duplicacy.DownloadConfig(self.storage, storagePassword)
	if err != nil {
		log.WithField("storageName", storageName).WithError(err).Fatal("failed to download config from storage")
	}

	if config == nil {
		log.WithField("storageName", storageName).Fatal("storage is not configured")
	}

	self.password = storagePassword
	self.config = config
	self.repository = repository

	// start background cleaner of our cache to keep memory use down
	go self.cleanReaddirCache(time.Minute * 2)
}