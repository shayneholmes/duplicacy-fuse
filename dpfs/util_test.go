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

func TestDpfs_abs(t *testing.T) {
	type args struct {
		filepath   string
		snapshotid string
		revision   int
	}
	tests := []struct {
		name             string
		self             *Dpfs
		args             args
		wantAbsolutepath string
	}{
		{"no snapshot id and no revision", &Dpfs{root: "snapshots"}, args{"/id/5/path", "", 0}, "snapshots/id/5/path"},
		{"snapshot id and no revision", &Dpfs{root: "snapshots"}, args{"/5/path", "id", 0}, "snapshots/id/5/path"},
		{"snapshot id and revision", &Dpfs{root: "snapshots"}, args{"/path", "id", 5}, "snapshots/id/5/path"},
		{"snapshot id root no revision", &Dpfs{root: "snapshots/id"}, args{"/5/path", "", 0}, "snapshots/id/5/path"},
		{"snapshot id root and revision", &Dpfs{root: "snapshots/id"}, args{"/path", "id", 5}, "snapshots/id/5/path"},
		{"revision root", &Dpfs{root: "snapshots/id/5"}, args{"/path", "id", 5}, "snapshots/id/5/path"},
		{"revision root subdir", &Dpfs{root: "snapshots/id/5"}, args{"/path/subdir", "id", 5}, "snapshots/id/5/path/subdir"},
	}
	for _, tt := range tests {
		abspath := tt.self.abs(tt.args.filepath, tt.args.snapshotid, tt.args.revision)
		assert.Equal(t, tt.wantAbsolutepath, abspath)
	}
}
