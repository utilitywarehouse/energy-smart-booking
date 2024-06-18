module github.com/utilitywarehouse/energy-smart-booking

go 1.22

toolchain go1.22.0

require (
	cloud.google.com/go/bigquery v1.61.0
	github.com/go-redis/redis/v8 v8.11.5
	github.com/golang/mock v1.6.0
	github.com/google/uuid v1.6.0
	github.com/gorilla/mux v1.8.1
	github.com/jackc/pgx/v5 v5.5.5
	github.com/json-iterator/go v1.1.12
	github.com/justinas/alice v1.2.0
	github.com/robfig/cron/v3 v3.0.1
	github.com/rubenv/sql-migrate v1.6.1
	github.com/sirupsen/logrus v1.9.3
	github.com/stretchr/testify v1.9.0
	github.com/urfave/cli/v2 v2.27.1
	github.com/utilitywarehouse/account-platform v0.0.0-20240416070355-de981e80aeb1
	github.com/utilitywarehouse/account-platform-protobuf-model v0.0.0-20240419140925-0e4ba8af6a21
	github.com/utilitywarehouse/click.uw.co.uk v0.0.0-20231012104247-b8d0609ca912
	github.com/utilitywarehouse/energy-contracts v1.155.1
	github.com/utilitywarehouse/energy-pkg/app v1.5.1
	github.com/utilitywarehouse/energy-pkg/fabrication v1.8.1
	github.com/utilitywarehouse/energy-pkg/grpc v0.1.8
	github.com/utilitywarehouse/energy-pkg/metrics v1.0.6
	github.com/utilitywarehouse/energy-pkg/ops v0.0.6
	github.com/utilitywarehouse/energy-pkg/postgres v0.3.7
	github.com/utilitywarehouse/energy-pkg/substratemessage v1.0.3
	github.com/utilitywarehouse/energy-pkg/substratemessage/v2 v2.2.5
	github.com/utilitywarehouse/go-ops-health-checks v1.0.1
	github.com/utilitywarehouse/go-ops-health-checks/v3 v3.1.0
	github.com/utilitywarehouse/uwos-go/iam v1.32.1
	github.com/utilitywarehouse/uwos-go/telemetry v1.28.3
	github.com/uw-labs/substrate v0.0.0-20240327161656-5cd769b67f2b
	github.com/uw-labs/substrate-tools v0.0.0-20210726101027-7ea25c77a95e
	golang.org/x/sync v0.7.0
	google.golang.org/api v0.180.0
	google.golang.org/genproto v0.0.0-20240604185151-ef581f913117
	google.golang.org/grpc v1.64.0
	google.golang.org/protobuf v1.34.1
)

require (
	cel.dev/expr v0.15.0 // indirect
	cloud.google.com/go/auth v0.4.1 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.2 // indirect
	cloud.google.com/go/iam v1.1.8 // indirect
	dario.cat/mergo v1.0.0 // indirect
	github.com/IBM/sarama v1.43.1 // indirect
	github.com/Microsoft/hcsshim v0.12.3 // indirect
	github.com/alvaroloes/enumer v1.1.2 // indirect
	github.com/apache/arrow/go/v15 v15.0.2 // indirect
	github.com/cloudflare/cfssl v1.6.5 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/google/certificate-transparency-go v1.1.8 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus v1.0.0 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware/v2 v2.1.0 // indirect
	github.com/jmoiron/sqlx v1.3.5 // indirect
	github.com/johanbrandhorst/certify v1.9.0 // indirect
	github.com/lmittmann/tint v1.0.4 // indirect
	github.com/lufia/plan9stats v0.0.0-20240408141607-282e7b5d6b74 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/moby/docker-image-spec v1.3.1 // indirect
	github.com/moby/sys/user v0.1.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/onsi/gomega v1.27.10 // indirect
	github.com/pascaldekloe/name v1.0.1 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/planetscale/vtprotobuf v0.6.0 // indirect
	github.com/power-devops/perfstat v0.0.0-20240221224432-82ca36839d55 // indirect
	github.com/shirou/gopsutil/v3 v3.24.3 // indirect
	github.com/shoenig/go-m1cpu v0.1.6 // indirect
	github.com/tklauser/go-sysconf v0.3.13 // indirect
	github.com/tklauser/numcpus v0.7.0 // indirect
	github.com/utilitywarehouse/protobuf-contracts v0.0.0-20240417111909-2b828f1813d1 // indirect
	github.com/utilitywarehouse/uwos-go/io v1.6.7 // indirect
	github.com/utilitywarehouse/uwos-go/runtime/k8sruntime v1.5.1 // indirect
	github.com/utilitywarehouse/uwos-go/storage/postgres v1.11.2 // indirect
	github.com/utilitywarehouse/uwos-go/time v1.4.3 // indirect
	github.com/uw-labs/proximo v0.0.0-20230125153035-a4cf3926a211 // indirect
	github.com/weppos/publicsuffix-go v0.30.2 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	github.com/zmap/zcrypto v0.0.0-20231219022726-a1f61fb1661c // indirect
	github.com/zmap/zlint/v3 v3.6.2 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.50.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/runtime v0.50.0 // indirect
	go.opentelemetry.io/contrib/propagators/autoprop v0.50.0 // indirect
	go.opentelemetry.io/contrib/propagators/aws v1.25.0 // indirect
	go.opentelemetry.io/contrib/propagators/b3 v1.25.0 // indirect
	go.opentelemetry.io/contrib/propagators/jaeger v1.25.0 // indirect
	go.opentelemetry.io/contrib/propagators/ot v1.25.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.25.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.25.0 // indirect
	go.opentelemetry.io/otel/exporters/prometheus v0.47.0 // indirect
	go.opentelemetry.io/otel/sdk v1.25.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.25.0 // indirect
	go.opentelemetry.io/proto/otlp v1.2.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/time v0.5.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240604185151-ef581f913117 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240604185151-ef581f913117 // indirect
)

require (
	cloud.google.com/go v0.114.0 // indirect
	cloud.google.com/go/compute/metadata v0.3.0 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20230124172434-306776ec8161 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/census-instrumentation/opencensus-proto v0.4.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cihub/seelog v0.0.0-20170130134532-f561c5e57575 // indirect
	github.com/cncf/xds/go v0.0.0-20240419173441-7e7c9eab8450 // indirect
	github.com/containerd/containerd v1.7.15 // indirect
	github.com/cpuguy83/dockercfg v0.3.1 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.4 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/docker/docker v26.1.0+incompatible // indirect
	github.com/docker/go-connections v0.5.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/eapache/go-resiliency v1.6.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20230731223053-c322873962e3 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/envoyproxy/go-control-plane v0.12.0 // indirect
	github.com/envoyproxy/protoc-gen-validate v1.0.4 // indirect
	github.com/franela/goreq v0.0.0-20171204163338-bcd34c9993f8 // indirect
	github.com/go-gorp/gorp/v3 v3.1.0 // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/gogo/protobuf v1.3.2
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/flatbuffers v24.3.25+incompatible // indirect
	github.com/google/go-cmp v0.6.0
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/s2a-go v0.1.7 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.2 // indirect
	github.com/googleapis/gax-go/v2 v2.12.4 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.4.0 // indirect
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.20.0
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-uuid v1.0.3 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20231201235250-de7065d80cb9 // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/jcmturner/aescts/v2 v2.0.0 // indirect
	github.com/jcmturner/dnsutils/v2 v2.0.0 // indirect
	github.com/jcmturner/gofork v1.7.6 // indirect
	github.com/jcmturner/gokrb5/v8 v8.4.4 // indirect
	github.com/jcmturner/rpc/v2 v2.0.3 // indirect
	github.com/klauspost/compress v1.17.8 // indirect
	github.com/klauspost/cpuid/v2 v2.2.7 // indirect
	github.com/lib/pq v1.10.9
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/moby/patternmatcher v0.6.0 // indirect
	github.com/moby/sys/sequential v0.5.0 // indirect
	github.com/moby/term v0.5.0 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0 // indirect
	github.com/opsgenie/opsgenie-go-sdk v0.0.0-20210305051615-8fb766c7514b // indirect
	github.com/pierrec/lz4/v4 v4.1.21 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.19.0
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.53.0 // indirect
	github.com/prometheus/procfs v0.14.0 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/testcontainers/testcontainers-go v0.30.0
	github.com/utilitywarehouse/energy-pkg/domain v1.30.0
	github.com/utilitywarehouse/energy-pkg/errorcodes v1.0.4 // indirect
	github.com/utilitywarehouse/go-operational v0.0.0-20220413104526-79ce40a50281
	github.com/utilitywarehouse/protoc-gen-uwentity v1.6.0 // indirect
	github.com/uw-labs/sync v0.0.0-20220413223303-ecb5d1fd966e // indirect
	github.com/xrash/smetrics v0.0.0-20240312152122-5f08fbb34913 // indirect
	github.com/zeebo/xxh3 v1.0.2 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.50.0 // indirect
	go.opentelemetry.io/otel v1.25.0
	go.opentelemetry.io/otel/metric v1.25.0 // indirect
	go.opentelemetry.io/otel/trace v1.25.0
	golang.org/x/crypto v0.24.0 // indirect
	golang.org/x/exp v0.0.0-20240416160154-fe59bbe5cc7f // indirect
	golang.org/x/mod v0.17.0 // indirect
	golang.org/x/net v0.26.0 // indirect
	golang.org/x/oauth2 v0.20.0 // indirect
	golang.org/x/sys v0.21.0 // indirect
	golang.org/x/text v0.16.0 // indirect
	golang.org/x/tools v0.21.1-0.20240508182429-e35e4ccd0d2d // indirect
	golang.org/x/xerrors v0.0.0-20231012003039-104605ab7028 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
