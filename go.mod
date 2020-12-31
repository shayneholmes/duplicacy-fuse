module gitlab.com/andrewheberle/duplicacy-fuse

replace gitlab.com/andrewheberle/duplicacy-fuse/dpfs => ./dpfs

go 1.13

require (
	cloud.google.com/go v0.52.0 // indirect
	github.com/billziss-gh/cgofuse v1.2.0
	github.com/gilbertchen/duplicacy v2.7.2+incompatible // indirect
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/klauspost/reedsolomon v1.9.10 // indirect
	github.com/minio/highwayhash v1.0.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pkg/xattr v0.4.3 // indirect
	github.com/sirupsen/logrus v1.4.2
	gitlab.com/andrewheberle/duplicacy-fuse/dpfs v0.0.0-00010101000000-000000000000
	go.opencensus.io v0.22.3 // indirect
	golang.org/x/exp v0.0.0-20200207192155-f17229e696bd // indirect
	golang.org/x/lint v0.0.0-20200130185559-910be7a94367 // indirect
	golang.org/x/tools v0.0.0-20200207224406-61798d64f025 // indirect
	google.golang.org/genproto v0.0.0-20200207204624-4f3edf09f4f6 // indirect
	google.golang.org/grpc v1.27.1 // indirect
)
