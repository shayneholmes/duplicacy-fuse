module gitlab.com/andrewheberle/duplicacy-fuse

replace gitlab.com/andrewheberle/duplicacy-fuse/dpfs => ./dpfs

go 1.26.0

require (
	github.com/billziss-gh/cgofuse v1.5.0
	github.com/sirupsen/logrus v1.10.2
	gitlab.com/andrewheberle/duplicacy-fuse/dpfs v0.0.0-20200211104317-8224be35869f
)

require (
	cel.dev/expr v0.25.3 // indirect
	cloud.google.com/go v0.123.0 // indirect
	cloud.google.com/go/auth v0.24.0 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.3.0 // indirect
	cloud.google.com/go/compute/metadata v0.10.0 // indirect
	cloud.google.com/go/iam v1.14.0 // indirect
	cloud.google.com/go/monitoring v1.31.0 // indirect
	cloud.google.com/go/storage v1.68.0 // indirect
	github.com/Azure/go-autorest v14.2.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest v0.11.30 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.9.24 // indirect
	github.com/Azure/go-autorest/autorest/date v0.3.1 // indirect
	github.com/Azure/go-autorest/logger v0.2.2 // indirect
	github.com/Azure/go-autorest/tracing v0.6.1 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/detectors/gcp v1.38.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric v0.62.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/internal/resourcemapping v0.62.0 // indirect
	github.com/OpenPeeDeeP/xdg v1.0.0 // indirect
	github.com/abcum/lcp v0.0.0-20201209214815-7a3f3840be81 // indirect
	github.com/aryann/difflib v0.0.0-20210328193216-ff5ff6dc229b // indirect
	github.com/aws/aws-sdk-go v1.55.8 // indirect
	github.com/bkaradzic/go-lz4 v1.0.0 // indirect
	github.com/calebcase/tmpfile v1.0.3 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cncf/xds/go v0.0.0-20260202195803-dba9d589def2 // indirect
	github.com/envoyproxy/go-control-plane/envoy v1.39.0 // indirect
	github.com/envoyproxy/protoc-gen-validate v1.3.3 // indirect
	github.com/felixge/httpsnoop v1.1.0 // indirect
	github.com/flynn/noise v1.0.0 // indirect
	github.com/geoffgarside/ber v1.1.0 // indirect
	github.com/gilbertchen/azure-sdk-for-go v14.1.2-0.20180323033227-8fd4663cab7c+incompatible // indirect
	github.com/gilbertchen/duplicacy v1.2.6-0.20250502193959-2def0161b377 // indirect
	github.com/gilbertchen/go-dropbox v0.0.0-20230321030224-087ef8db1916 // indirect
	github.com/gilbertchen/go-ole v1.2.0 // indirect
	github.com/gilbertchen/goamz v0.0.0-20170712012135-eada9f4e8cc2 // indirect
	github.com/gilbertchen/gopass v0.0.0-20170109162249-bf9dde6d0d2c // indirect
	github.com/gilbertchen/highwayhash v0.0.0-20221109044721-eeab1f4799d8 // indirect
	github.com/gilbertchen/keyring v0.0.0-20221004152639-1661cbebc508 // indirect
	github.com/go-jose/go-jose/v4 v4.1.5 // indirect
	github.com/go-logr/logr v1.4.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/godbus/dbus v4.1.0+incompatible // indirect
	github.com/gofrs/flock v0.8.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.2 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/s2a-go v0.1.10 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.22 // indirect
	github.com/googleapis/gax-go/v2 v2.26.2 // indirect
	github.com/hashicorp/go-immutable-radix/v2 v2.0.0 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/hirochachacha/go-smb2 v1.1.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/jtolio/eventkit v0.0.0-20221004135224-074cf276595b // indirect
	github.com/jtolio/noiseconn v0.0.0-20230111204749-d7ec1a08b0b8 // indirect
	github.com/klauspost/compress v1.16.3 // indirect
	github.com/klauspost/cpuid/v2 v2.4.0 // indirect
	github.com/klauspost/reedsolomon v1.14.2 // indirect
	github.com/kr/fs v0.1.0 // indirect
	github.com/marstr/guid v1.1.0 // indirect
	github.com/mattetti/filebuffer v1.0.1 // indirect
	github.com/minio/blake2b-simd v0.0.0-20160723061019-3f5f724cb5b1 // indirect
	github.com/minio/highwayhash v1.0.4 // indirect
	github.com/ncw/swift/v2 v2.0.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pkg/sftp v1.13.11 // indirect
	github.com/pkg/xattr v0.4.12 // indirect
	github.com/planetscale/vtprotobuf v0.6.1-0.20240319094008-0393e58bdf10 // indirect
	github.com/satori/go.uuid v1.2.0 // indirect
	github.com/spacemonkeygo/monkit/v3 v3.0.22 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/spiffe/go-spiffe/v2 v2.8.2 // indirect
	github.com/vaughan0/go-ini v0.0.0-20130923145212-a98ad7ee00ec // indirect
	github.com/vivint/infectious v0.0.0-20200605153912-25a574ae18a3 // indirect
	github.com/vmihailenco/msgpack v4.0.4+incompatible // indirect
	github.com/zeebo/blake3 v0.2.3 // indirect
	github.com/zeebo/errs v1.4.0 // indirect
	go.mills.io/bitcask/v2 v2.2.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/contrib/detectors/gcp v1.46.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.71.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.71.0 // indirect
	go.opentelemetry.io/otel v1.46.0 // indirect
	go.opentelemetry.io/otel/metric v1.46.0 // indirect
	go.opentelemetry.io/otel/sdk v1.46.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.46.0 // indirect
	go.opentelemetry.io/otel/trace v1.46.0 // indirect
	golang.org/x/crypto v0.57.0 // indirect
	golang.org/x/exp v0.0.0-20260908205506-85c1c2202aba // indirect
	golang.org/x/net v0.59.0 // indirect
	golang.org/x/oauth2 v0.37.0 // indirect
	golang.org/x/sync v0.23.0 // indirect
	golang.org/x/sys v0.48.0 // indirect
	golang.org/x/term v0.46.0 // indirect
	golang.org/x/text v0.42.0 // indirect
	golang.org/x/time v0.16.0 // indirect
	google.golang.org/api v0.299.0 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/genproto v0.0.0-20260921155816-b14227669459 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260921155816-b14227669459 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260921155816-b14227669459 // indirect
	google.golang.org/grpc v1.84.0 // indirect
	google.golang.org/protobuf v1.36.12 // indirect
	gopkg.in/yaml.v2 v2.3.0 // indirect
	storj.io/common v0.0.0-20230920095429-0ce0a575e6f8 // indirect
	storj.io/drpc v0.0.33 // indirect
	storj.io/picobuf v0.0.2-0.20230906122608-c4ba17033c6c // indirect
	storj.io/uplink v1.12.1 // indirect
)
