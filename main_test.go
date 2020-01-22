package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDpfs_revision(t *testing.T) {
	allDpfs := Dpfs{root: "snapshots"}
	idDpfs := Dpfs{root: "snapshots/id"}
	tests := []struct {
		name string
		self *Dpfs
		p    string
		want string
	}{
		{"all", &allDpfs, "/", ""},
		{"all", &allDpfs, "/id", ""},
		{"all", &allDpfs, "/id/3", "3"},
		{"id", &idDpfs, "/", ""},
		{"id", &idDpfs, "/3", "3"},
	}
	for _, tt := range tests {
		got := tt.self.revision(tt.p)
		assert.Equal(t, tt.want, got, fmt.Sprintf("%s - %s", tt.name, tt.p))
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
		{&allDpfs, "snapshots", "", 0, "", false},
		{&allDpfs, "snapshots/id", "id", 0, "", false},
		{&allDpfs, "snapshots/id/3", "id", 3, "", false},
		{&allDpfs, "snapshots/id/3/filename.txt", "id", 3, "filename.txt", false},
		{&allDpfs, "snapshots/id/X", "id", 0, "", true},
		{&allDpfs, "snapshots/id/X/filename.txt", "id", 0, "filename.txt", true},
		{&idDpfs, "/", "id", 0, "", false},
		{&idDpfs, "/3", "id", 3, "", false},
		{&idDpfs, "/3/filename.txt", "id", 3, "filename.txt", false},
		{&idDpfs, "/X", "id", 0, "", true},
		{&idDpfs, "/X/filename.txt", "id", 0, "filename.txt", true},
		{&idDpfs, "snapshots/id", "id", 0, "", false},
		{&idDpfs, "snapshots/id/3", "id", 3, "", false},
		{&idDpfs, "snapshots/id/3/filename.txt", "id", 3, "filename.txt", false},
		{&idDpfs, "snapshots/id/X", "id", 0, "", true},
		{&idDpfs, "snapshots/id/X/filename.txt", "id", 0, "filename.txt", true},
		{&revDpfs, "/", "id", 3, "", false},
		{&revDpfs, "/4", "id", 3, "4", false},
		{&revDpfs, "/4/filename.txt", "id", 3, "4/filename.txt", false},
		{&revDpfs, "/X", "id", 3, "X", false},
		{&revDpfs, "/X/filename.txt", "id", 3, "X/filename.txt", false},
		{&revDpfs, "snapshots/id/3", "id", 3, "", false},
		{&revDpfs, "snapshots/id/3/filename.txt", "id", 3, "filename.txt", false},
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

func TestDpfs_snapshotPath(t *testing.T) {
	allDpfs := Dpfs{root: "snapshots"}
	idDpfs := Dpfs{root: "snapshots/id"}
	revDpfs := Dpfs{root: "snapshots/id/3"}
	tests := []struct {
		self *Dpfs
		p    string
		want string
	}{
		{&allDpfs, "/", "snapshots"},
		{&allDpfs, "/id", "snapshots/id"},
		{&allDpfs, "/id/3", "snapshots/id/3"},
		{&allDpfs, "/id/3/filename.txt", "snapshots/id/3/filename.txt"},
		{&allDpfs, "snapshots", "snapshots"},
		{&allDpfs, "snapshots/id", "snapshots/id"},
		{&allDpfs, "snapshots/id/3", "snapshots/id/3"},
		{&allDpfs, "snapshots/id/3/filename.txt", "snapshots/id/3/filename.txt"},
		{&idDpfs, "/", "snapshots/id"},
		{&idDpfs, "/3", "snapshots/id/3"},
		{&idDpfs, "/3/filename.txt", "snapshots/id/3/filename.txt"},
		{&idDpfs, "snapshots/id", "snapshots/id"},
		{&idDpfs, "snapshots/id/3", "snapshots/id/3"},
		{&idDpfs, "snapshots/id/3/filename.txt", "snapshots/id/3/filename.txt"},
		{&revDpfs, "/", "snapshots/id/3"},
		{&revDpfs, "/filename.txt", "snapshots/id/3/filename.txt"},
		{&revDpfs, "snapshots/id/3", "snapshots/id/3"},
		{&revDpfs, "snapshots/id/3/filename.txt", "snapshots/id/3/filename.txt"},
	}
	for _, tt := range tests {
		got := tt.self.snapshotPath(tt.p)
		assert.Equal(t, tt.want, got)
	}
}
