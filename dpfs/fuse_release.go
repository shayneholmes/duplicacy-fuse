package dpfs

// Release is a noop as we are not worried about saving any changes
func (self *Dpfs) Release(path string, fh uint64) (errc int) {
	return 0
}
