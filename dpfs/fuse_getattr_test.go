package dpfs

import (
	"testing"

	"github.com/billziss-gh/cgofuse/fuse"
	"github.com/stretchr/testify/assert"
)

func TestDpfs_Getattr(t *testing.T) {
	tests := []struct {
		self *Dpfs
		path string
		stat *fuse.Stat_t
		fh   uint64
		errc int
	}{
		{&Dpfs{}, "/desktop.ini", &fuse.Stat_t{}, 0, -fuse.ENOENT},
		{&Dpfs{}, "/id/desktop.ini", &fuse.Stat_t{}, 0, -fuse.ENOENT},
		{&Dpfs{snapshotid: "id"}, "/desktop.ini", &fuse.Stat_t{}, 0, -fuse.ENOENT},
	}

	for _, tt := range tests {
		errc := tt.self.Getattr(tt.path, tt.stat, tt.fh)
		assert.Equal(t, tt.errc, errc)
	}
}
