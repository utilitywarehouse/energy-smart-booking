module github.com/utilitywarehouse/energy-smart-booking

go 1.24.1

require (
	cloud.google.com/go/bigquery v1.67.0
	github.com/go-redis/redis/v8 v8.11.5
	github.com/gogo/protobuf v1.3.2
	github.com/google/go-cmp v0.7.0
	github.com/google/uuid v1.6.0
	github.com/gorilla/mux v1.8.1
	github.com/jackc/pgx/v5 v5.7.4
	github.com/json-iterator/go v1.1.12
	github.com/justinas/alice v1.2.0
	github.com/lib/pq v1.10.9
	github.com/robfig/cron/v3 v3.0.1
	github.com/rubenv/sql-migrate v1.7.1
	github.com/sirupsen/logrus v1.9.3
	github.com/stretchr/testify v1.10.0
	github.com/testcontainers/testcontainers-go v0.36.0
	github.com/urfave/cli/v2 v2.27.6
	github.com/utilitywarehouse/account-platform v0.0.0-20250404133623-1740762b394c
	github.com/utilitywarehouse/account-platform-protobuf-model v0.0.0-20250402131320-0d4e0d75d0b2
	github.com/utilitywarehouse/bill-contracts v1.0.1
	github.com/utilitywarehouse/click.uw.co.uk v0.0.0-20231012104247-b8d0609ca912
	github.com/utilitywarehouse/energy-contracts v1.229.0
	github.com/utilitywarehouse/energy-pkg/app v1.6.1
	github.com/utilitywarehouse/energy-pkg/domain v1.34.2
	github.com/utilitywarehouse/energy-pkg/fabrication v1.8.2
	github.com/utilitywarehouse/energy-pkg/grpc v0.1.9
	github.com/utilitywarehouse/energy-pkg/metrics v1.0.6
	github.com/utilitywarehouse/energy-pkg/ops v0.0.9
	github.com/utilitywarehouse/energy-pkg/postgres v0.3.9
	github.com/utilitywarehouse/energy-pkg/substratemessage v1.0.3
	github.com/utilitywarehouse/energy-pkg/substratemessage/v2 v2.2.6
	github.com/utilitywarehouse/go-ops-health-checks v1.0.1
	github.com/utilitywarehouse/go-ops-health-checks/v3 v3.1.0
	github.com/utilitywarehouse/uwos-go/iam v1.40.0
	github.com/utilitywarehouse/uwos-go/telemetry v1.38.1
	github.com/utilitywarehouse/uwos-go/x/bill v1.1.5
	github.com/uw-labs/substrate v0.0.0-20240327161656-5cd769b67f2b
	github.com/uw-labs/substrate-tools v0.0.0-20210726101027-7ea25c77a95e
	go.opentelemetry.io/otel v1.35.0
	go.opentelemetry.io/otel/trace v1.35.0
	go.uber.org/mock v0.5.0
	golang.org/x/sync v0.12.0
	google.golang.org/api v0.228.0
	google.golang.org/genproto v0.0.0-20250404141209-ee84b53bf3d0
	google.golang.org/grpc v1.71.1
	google.golang.org/protobuf v1.36.6
)

require (
	buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go v1.36.5-20250219170025-d39267d9df8f.1 // indirect
	cel.dev/expr v0.23.0 // indirect
	cloud.google.com/go/auth v0.15.0 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.8 // indirect
	cloud.google.com/go/iam v1.5.0 // indirect
	dario.cat/mergo v1.0.1 // indirect
	github.com/IBM/sarama v1.45.1 // indirect
	github.com/alvaroloes/enumer v1.1.2 // indirect
	github.com/antlr4-go/antlr/v4 v4.13.1 // indirect
	github.com/apache/arrow/go/v15 v15.0.2 // indirect
	github.com/bufbuild/protovalidate-go v0.9.2 // indirect
	github.com/cloudflare/cfssl v1.6.5 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/containerd/platforms v0.2.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/ebitengine/purego v0.8.2 // indirect
	github.com/envoyproxy/go-control-plane/envoy v1.32.4 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/google/cel-go v0.23.2 // indirect
	github.com/google/certificate-transparency-go v1.3.1 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus v1.0.1 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware/v2 v2.3.1 // indirect
	github.com/jmoiron/sqlx v1.4.0 // indirect
	github.com/johanbrandhorst/certify v1.9.0 // indirect
	github.com/lmittmann/tint v1.0.7 // indirect
	github.com/lufia/plan9stats v0.0.0-20240909124753-873cd0166683 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/moby/docker-image-spec v1.3.1 // indirect
	github.com/moby/sys/user v0.4.0 // indirect
	github.com/moby/sys/userns v0.1.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/pascaldekloe/name v1.0.1 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/planetscale/vtprotobuf v0.6.1-0.20240319094008-0393e58bdf10 // indirect
	github.com/power-devops/perfstat v0.0.0-20240221224432-82ca36839d55 // indirect
	github.com/shirou/gopsutil/v4 v4.25.1 // indirect
	github.com/stoewer/go-strcase v1.3.0 // indirect
	github.com/tklauser/go-sysconf v0.3.14 // indirect
	github.com/tklauser/numcpus v0.8.0 // indirect
	github.com/twmb/franz-go v1.18.1 // indirect
	github.com/twmb/franz-go/pkg/kmsg v1.9.0 // indirect
	github.com/twmb/franz-go/plugin/kotel v1.5.0 // indirect
	github.com/twmb/franz-go/plugin/kslog v1.0.0 // indirect
	github.com/utilitywarehouse/protobuf-contracts v0.0.0-20250404083004-f2d11cb9368c // indirect
	github.com/utilitywarehouse/uwos-go/crypto/tlsconfig v1.4.4 // indirect
	github.com/utilitywarehouse/uwos-go/grpc v1.9.0 // indirect
	github.com/utilitywarehouse/uwos-go/io v1.6.22 // indirect
	github.com/utilitywarehouse/uwos-go/pubsub/kafka v1.6.9 // indirect
	github.com/utilitywarehouse/uwos-go/runtime/k8sruntime v1.5.3 // indirect
	github.com/utilitywarehouse/uwos-go/storage/postgres v1.12.9 // indirect
	github.com/utilitywarehouse/uwos-go/time v1.4.12 // indirect
	github.com/utilitywarehouse/uwos-go/x/build v1.3.1 // indirect
	github.com/uw-labs/proximo v0.0.0-20230125153035-a4cf3926a211 // indirect
	github.com/weppos/publicsuffix-go v0.40.3-0.20250127173806-e489a31678ca // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	github.com/zmap/zcrypto v0.0.0-20250129210703-03c45d0bae98 // indirect
	github.com/zmap/zlint/v3 v3.6.5 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.59.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/runtime v0.60.0 // indirect
	go.opentelemetry.io/contrib/propagators/autoprop v0.60.0 // indirect
	go.opentelemetry.io/contrib/propagators/aws v1.35.0 // indirect
	go.opentelemetry.io/contrib/propagators/b3 v1.35.0 // indirect
	go.opentelemetry.io/contrib/propagators/jaeger v1.35.0 // indirect
	go.opentelemetry.io/contrib/propagators/ot v1.35.0 // indirect
	go.opentelemetry.io/contrib/samplers/jaegerremote v0.29.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.35.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.35.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.35.0 // indirect
	go.opentelemetry.io/otel/exporters/prometheus v0.57.0 // indirect
	go.opentelemetry.io/otel/sdk v1.35.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.35.0 // indirect
	go.opentelemetry.io/proto/otlp v1.5.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/time v0.11.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250324211829-b45e905df463 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250324211829-b45e905df463 // indirect
	tlog.app/go/loc v0.7.2 // indirect
)

require (
	cloud.google.com/go v0.120.0 // indirect
	cloud.google.com/go/compute/metadata v0.6.0 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20250102033503-faa5f7b0171c // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cihub/seelog v0.0.0-20170130134532-f561c5e57575 // indirect
	github.com/cncf/xds/go v0.0.0-20250121191232-2f005788dc42 // indirect
	github.com/cpuguy83/dockercfg v0.3.2 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.6 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/docker/docker v28.0.4+incompatible // indirect
	github.com/docker/go-connections v0.5.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/eapache/go-resiliency v1.7.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20230731223053-c322873962e3 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/envoyproxy/protoc-gen-validate v1.2.1 // indirect
	github.com/franela/goreq v0.0.0-20171204163338-bcd34c9993f8 // indirect
	github.com/go-gorp/gorp/v3 v3.1.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/flatbuffers v24.3.25+incompatible // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/s2a-go v0.1.9 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.6 // indirect
	github.com/googleapis/gax-go/v2 v2.14.1 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.4.0 // indirect
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.26.3
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-uuid v1.0.3 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jcmturner/aescts/v2 v2.0.0 // indirect
	github.com/jcmturner/dnsutils/v2 v2.0.0 // indirect
	github.com/jcmturner/gofork v1.7.6 // indirect
	github.com/jcmturner/gokrb5/v8 v8.4.4 // indirect
	github.com/jcmturner/rpc/v2 v2.0.3 // indirect
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/klauspost/cpuid/v2 v2.2.7 // indirect
	github.com/magiconair/properties v1.8.9 // indirect
	github.com/moby/patternmatcher v0.6.0 // indirect
	github.com/moby/sys/sequential v0.6.0 // indirect
	github.com/moby/term v0.5.2 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.1 // indirect
	github.com/opsgenie/opsgenie-go-sdk v0.0.0-20210305051615-8fb766c7514b // indirect
	github.com/pierrec/lz4/v4 v4.1.22 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.21.1
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.62.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/utilitywarehouse/energy-pkg/errorcodes v1.0.4 // indirect
	github.com/utilitywarehouse/go-operational v0.0.0-20250206100814-e7d65e48b364
	github.com/utilitywarehouse/protoc-gen-uwentity v1.6.0 // indirect
	github.com/uw-labs/sync v0.0.0-20220413223303-ecb5d1fd966e // indirect
	github.com/xrash/smetrics v0.0.0-20240521201337-686a1a2994c1 // indirect
	github.com/zeebo/xxh3 v1.0.2 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.60.0 // indirect
	go.opentelemetry.io/otel/metric v1.35.0 // indirect
	golang.org/x/crypto v0.36.0 // indirect
	golang.org/x/exp v0.0.0-20250305212735-054e65f0b394 // indirect
	golang.org/x/mod v0.24.0 // indirect
	golang.org/x/net v0.38.0 // indirect
	golang.org/x/oauth2 v0.28.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	golang.org/x/tools v0.31.0 // indirect
	golang.org/x/xerrors v0.0.0-20240903120638-7835f813f4da // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
