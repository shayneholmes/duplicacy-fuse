# duplicacy-fuse

## Overview

This is my first effort at implementing a file system in userspace.

## Purpose

Allows snapshots and revisions on storage in a configured duplicacy repository to be mounted (read-only) as a file system.

## Building

Should build and run on both Windows and Linux.

On Windows, it can be built with, or without CGO support using the CGO_ENABLED env var.

## Requirements

Check out [cgofuse](https://github.com/billziss-gh/cgofuse) for full requriments, however requirements are:

* Linux : gcc, libfuse-dev
* Windows (cgo) : [WinFsp](https://github.com/billziss-gh/winfsp), gcc
* Windows (!cgo)  : [WinFsp](https://github.com/billziss-gh/winfsp)

## Usage

```sh
duplicacy-fuse <mount point> [options]
```

### Options

  `-o repositorty=<path>` : Path to repository. Defaults to current directory.
  
  `-o storage=<storage name>` : Remote storage to mount. Defaults to "default".

  `-o password=<storage password>`: Password for the specified storage.
  
  `-o snapshot=<id>` : Which snapshot id to mount. Unless `all` option is specified, defaults to the repository snapshot if for the chosen storage.
  
  `-o revision=<number>` : Which revision to mount. Cannot be specified when the `all` is specified.
  
  `-o all` : Display all snapshot ids for the specified storage.
  
## Status

Basic functionality seems to all work fine, although this has only been tested on Windows.

Performance is quite slow, especially when opening files.

Memory use is quite high also as the listing of all files in a snapshot is loaded into memory when initially browsing a revision.

Browsing files and directories is also quite inneficient as the list of files is searched every time to find matching files, so worst case N-1 entries may need to be searched to get file info for a single file.

Opening files is also very slow, but this may be due to using an innefficient process to dowload file contents.