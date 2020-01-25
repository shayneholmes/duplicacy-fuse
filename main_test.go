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
