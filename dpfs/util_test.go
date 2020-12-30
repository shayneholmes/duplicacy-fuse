package dpfs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDpfs_newpathInfo(t *testing.T) {
	tests := []struct {
		name     string
		self     *Dpfs
		filepath string
		want     pathInfo
		str      string
	}{
		{"&Dpfs{}", &Dpfs{}, "/", pathInfo{}, "snapshots"},
		{"&Dpfs{}", &Dpfs{}, "/desktop.ini", pathInfo{snapshotid: "desktop.ini"}, "snapshots/desktop.ini"},
		{"&Dpfs{}", &Dpfs{}, "/id/desktop.ini", pathInfo{snapshotid: "id"}, "snapshots/id"},
		{"&Dpfs{}", &Dpfs{}, "/id", pathInfo{snapshotid: "id"}, "snapshots/id"},
		{"&Dpfs{}", &Dpfs{}, "/id/3", pathInfo{snapshotid: "id", revision: 3}, "snapshots/id/3"},
		{"&Dpfs{}", &Dpfs{}, "/id/3/file.txt", pathInfo{snapshotid: "id", revision: 3, filepath: "/file.txt"}, "snapshots/id/3/file.txt"},
		{"&Dpfs{}", &Dpfs{}, "/id/3/dir/file.txt", pathInfo{snapshotid: "id", revision: 3, filepath: "/dir/file.txt"}, "snapshots/id/3/dir/file.txt"},
		{"&Dpfs{snapshotid: \"id\"}", &Dpfs{snapshotid: "id"}, "/", pathInfo{snapshotid: "id"}, "snapshots/id"},
		{"&Dpfs{snapshotid: \"id\"}", &Dpfs{snapshotid: "id"}, "/3", pathInfo{snapshotid: "id", revision: 3}, "snapshots/id/3"},
		{"&Dpfs{snapshotid: \"id\"}", &Dpfs{snapshotid: "id"}, "/3/file.txt", pathInfo{snapshotid: "id", revision: 3, filepath: "/file.txt"}, "snapshots/id/3/file.txt"},
		{"&Dpfs{snapshotid: \"id\"}", &Dpfs{snapshotid: "id"}, "/3/dir/file.txt", pathInfo{snapshotid: "id", revision: 3, filepath: "/dir/file.txt"}, "snapshots/id/3/dir/file.txt"},
		{"&Dpfs{snapshotid: \"id\", revision: 3}", &Dpfs{snapshotid: "id", revision: 3}, "/", pathInfo{snapshotid: "id", revision: 3, filepath: "/"}, "snapshots/id/3/"},
		{"&Dpfs{snapshotid: \"id\", revision: 3}", &Dpfs{snapshotid: "id", revision: 3}, "/file.txt", pathInfo{snapshotid: "id", revision: 3, filepath: "/file.txt"}, "snapshots/id/3/file.txt"},
		{"&Dpfs{snapshotid: \"id\", revision: 3}", &Dpfs{snapshotid: "id", revision: 3}, "/dir/file.txt", pathInfo{snapshotid: "id", revision: 3, filepath: "/dir/file.txt"}, "snapshots/id/3/dir/file.txt"},
	}
	for _, tt := range tests {
		got := tt.self.newpathInfo(tt.filepath)
		assert.Equal(t, tt.want, got, tt.name)
		assert.Equal(t, tt.str, got.String(), tt.name)
	}
}
