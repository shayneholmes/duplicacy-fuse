package dpfs

import (
	log "github.com/sirupsen/logrus"
)

// Access satisfies the Access implementation from fuse.FileSystemInterface
func (self *Dpfs) Access(path string, mask uint32) int {
	log.WithField("operation", "Access").Warning("not implemented")
	return NotImplemented
}
