package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDpfs_snapshotPath(t *testing.T) {
	allDpfs := Dpfs{root: "snapshots"}
	idDpfs := Dpfs{root: "snapshots", snapshotid: "id"}
	revDpfs := Dpfs{root: "snapshots", snapshotid: "id", revision: 3}
	tests := []struct {
		self *Dpfs
		p    string
		want string
	}{
		{&allDpfs, "/", "snapshots"},
		{&allDpfs, "/id", "snapshots/id"},
		{&allDpfs, "/id/3", "snapshots/id/3"},
		{&allDpfs, "/id/3/filename.txt", "snapshots/id/3/filename.txt"},
		{&idDpfs, "/", "snapshots/id"},
		{&idDpfs, "/3", "snapshots/id/3"},
		{&idDpfs, "/3/filename.txt", "snapshots/id/3/filename.txt"},
		{&revDpfs, "/", "snapshots/id/3"},
		{&revDpfs, "/filename.txt", "snapshots/id/3/filename.txt"},
	}
	for _, tt := range tests {
		got := tt.self.snapshotPath(tt.p)
		assert.Equal(t, tt.want, got)
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

func TestDpfs_info(t *testing.T) {
	allDpfs := Dpfs{root: "snapshots"}
	idDpfs := Dpfs{root: "snapshots/id"}
	revDpfs := Dpfs{root: "snapshots/id/3"}
	tests := []struct {
		self           *Dpfs
		p              string
		wantSnapshotid string
		wantRevision   int
		wantPath       string
		wantErr        bool
	}{
		{&allDpfs, "/", "", 0, "", false},
		{&allDpfs, "/id", "id", 0, "", false},
		{&allDpfs, "/id/3", "id", 3, "", false},
		{&allDpfs, "/id/3/filename.txt", "id", 3, "filename.txt", false},
		{&allDpfs, "/id/X", "id", 0, "", true},
		{&allDpfs, "/id/X/filename.txt", "id", 0, "filename.txt", true},
		{&idDpfs, "/", "id", 0, "", false},
		{&idDpfs, "/3", "id", 3, "", false},
		{&idDpfs, "/3/filename.txt", "id", 3, "filename.txt", false},
		{&idDpfs, "/X", "id", 0, "", true},
		{&idDpfs, "/X/filename.txt", "id", 0, "filename.txt", true},
		{&revDpfs, "/", "id", 3, "", false},
		{&revDpfs, "/4", "id", 3, "4", false},
		{&revDpfs, "/4/filename.txt", "id", 3, "4/filename.txt", false},
		{&revDpfs, "/X", "id", 3, "X", false},
		{&revDpfs, "/X/filename.txt", "id", 3, "X/filename.txt", false},
	}
	for _, tt := range tests {
		name := fmt.Sprintf("root: %s; p: %s; snapshotid: %s; revision: %d; path: %s; err: %v", tt.self.root, tt.p, tt.wantSnapshotid, tt.wantRevision, tt.wantPath, tt.wantErr)
		gotSnapshotid, gotRevision, gotPath, gotErr := tt.self.info(tt.p)
		if tt.wantErr {
			assert.NotNil(t, gotErr, name)
			continue
		}
		if assert.Nil(t, gotErr) {
			assert.Equal(t, tt.wantSnapshotid, gotSnapshotid, name)
			assert.Equal(t, tt.wantRevision, gotRevision, name)
			assert.Equal(t, tt.wantPath, gotPath, name)
		}

	}
}
