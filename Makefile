.PHONY : all

all : duplicacy-fuse duplicacy-fuse.exe

windows : duplicacy-fuse.exe

linux : duplicacy-fuse

duplicacy-fuse : *.go
	env GOOS=linux go build .

duplicacy-fuse.exe : *.go
	env GOOS=windows go build .
