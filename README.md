# duplicacy-fuse

This is my first effort at implementing a file system in userspace.

## Purpose

Allows snapshots and revisions on storage in a configured duplicacy repository to be mounted (read-only) as a file system.

## Building

Should build and run on both Windows and Linux.

On Windows, it can be built with, or without CGO support using the GOCGO env var.

## Requirements

On Linux you will need fuse support and on Windows you need ... installed.

## Usage

'''sh
duplicacy-fuse <mount point> [options]
'''

### Options

  -o repositorty=<path> : Path to repository. Defaults to current directory.
  
  -o storage=<storage name> :

