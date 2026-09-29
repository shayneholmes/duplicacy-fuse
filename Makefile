.PHONY : all

all : duplicacy-fuse

darwin : duplicacy-fuse

duplicacy-fuse : *.go dpfs/*.go
	env GOOS=darwin asdf exec go build .
