package dpfs

import (
	log "github.com/sirupsen/logrus"
)

// Destroy satisfies the Destroy implementation from fuse.FileSystemInterface
func (self *Dpfs) Destroy() {
    if err := self.cache.Close(); err != nil {
        log.WithError(err).Info()
    }
}