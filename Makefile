.PHONY : all

all : duplicacy-fuse

darwin : duplicacy-fuse

duplicacy-fuse : *.go
	env GOOS=darwin go build .
