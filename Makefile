.PHONY : all

all : duplicacy-fuse duplicacy-fuse.exe

windows : duplicacy-fuse.exe

darwin : duplicacy-fuse

duplicacy-fuse : *.go
	env GOOS=darwin go build .

duplicacy-fuse.exe : *.go
	env GOOS=windows go build .
