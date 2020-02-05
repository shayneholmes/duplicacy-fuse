package main

import (
	"os"
	"path"
	"time"

	"github.com/billziss-gh/cgofuse/fuse"
	duplicacy "github.com/gilbertchen/duplicacy/src"
	log "github.com/sirupsen/logrus"
)

func (self *Dpfs) Init() {
	var repository, storageName, storagePassword, loglevel string
	var snapshot int
	var debug, all bool

	self.lock.Lock()
	defer self.lock.Unlock()

	_, err := fuse.OptParse(os.Args, "repository=%s storage=%s snapshot=%d password=%s loglevel=%s debug all", &repository, &storageName, &snapshot, &storagePassword, &loglevel, &debug, &all)
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

	if storageName == "" {
		storageName = "default"
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

	if all {
		self.root = "snapshots"
	} else {
		self.root = path.Join("snapshots", self.preference.SnapshotID)
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
	self.ofiles = make(map[uint64]node_t)

	go self.cleanReaddirCache(time.Minute * 2)
}
