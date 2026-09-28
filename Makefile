.PHONY : all

all : duplicacy-fuse

darwin : duplicacy-fuse

duplicacy-fuse : *.go dpfs/*.go
	env GOOS=darwin go1.26.8 build .
