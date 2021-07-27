module github.com/fluxcd/helm-operator

go 1.14

require (
	github.com/bugsnag/panicwrap v1.2.0 // indirect
	github.com/docker/go-metrics v0.0.1 // indirect
	github.com/docker/libtrust v0.0.0-20160708172513-aabc10ec26b7 // indirect
	github.com/fatih/color v1.9.0 // indirect
	github.com/fluxcd/flux v1.17.2-0.20200121140732-3903cf8e71c3
	github.com/fluxcd/helm-operator/pkg/install v0.0.0-00010101000000-000000000000
	github.com/garyburd/redigo v1.6.0 // indirect
	github.com/go-kit/kit v0.10.0
	github.com/gofrs/uuid v3.2.0+incompatible // indirect
	github.com/golang/protobuf v1.5.0
	github.com/google/go-cmp v0.5.5
	github.com/gorilla/mux v1.7.3
	github.com/helm/helm-2to3 v0.8.2
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0 // indirect
	github.com/mattn/go-colorable v0.1.6 // indirect
	github.com/ncabatoff/go-seq v0.0.0-20180805175032-b08ef85ed833
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.7.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	github.com/yvasiyarov/go-metrics v0.0.0-20150112132944-c25f46c4b940 // indirect
	github.com/yvasiyarov/newrelic_platform_go v0.0.0-20160601141957-9c099fbc30e9 // indirect
	google.golang.org/grpc v1.27.1
	helm.sh/helm/v3 v3.6.3
	k8s.io/api v0.21.7
	k8s.io/apiextensions-apiserver v0.21.7
	k8s.io/apimachinery v0.21.7
	k8s.io/cli-runtime v0.21.7
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/helm v2.17.0+incompatible
	k8s.io/klog v1.0.0
	k8s.io/kubectl v0.21.7
	k8s.io/utils v0.0.0-20210521133846-da695404a2bc
	sigs.k8s.io/yaml v1.2.0
)

// github.com/fluxcd/helm-operator/pkg/install lives in this very reprository, so use that
replace github.com/fluxcd/helm-operator/pkg/install => ./pkg/install

// Transitive requirement from Helm: https://github.com/helm/helm/blob/v3.5.2/go.mod#L52-L53
replace (
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d
	github.com/docker/docker => github.com/moby/moby v17.12.0-ce-rc1.0.20200618181300-9dc6525e6118+incompatible
)

// Pin Flux to 1.18.0
replace (
	github.com/fluxcd/flux => github.com/fluxcd/flux v1.18.0
	github.com/fluxcd/flux/pkg/install => github.com/fluxcd/flux/pkg/install v0.0.0-20200206191601-8b676b003ab0
)

// Force upgrade because of a transitive downgrade.
// github.com/fluxcd/helm-operator
// +-> github.com/fluxcd/flux@v1.17.2
//     +-> k8s.io/client-go@v11.0.0+incompatible
replace k8s.io/client-go => k8s.io/client-go v0.21.7

// Force upgrade because of a transitive downgrade.
// github.com/fluxcd/flux
// +-> github.com/fluxcd/helm-operator@v1.0.0-rc6
//     +-> helm.sh/helm/v3@v3.1.2
//     +-> helm.sh/helm@v2.16.1
replace (
	helm.sh/helm/v3 => helm.sh/helm/v3 v3.6.3
	k8s.io/helm => k8s.io/helm v2.17.0+incompatible
)

// Pin specific runc following example of source-controller upstream:
// https://github.com/fluxcd/source-controller/blob/59dc602b2d0a5b0d4625a4a848303fa2d2117e84/go.mod#L62-L77
// github.com/fluxcd/helm-operator
// +-> helm.sh/helm/v3@v3.6.3
//     +-> github.com/opencontainers/runc@v1.0.2
replace github.com/opencontainers/runc => github.com/opencontainers/runc v1.0.2

// Mitigating: GHSA-5j5w-g665-5m35
replace github.com/containerd/containerd => github.com/containerd/containerd v1.4.12

// Mitigating: GHSA-77vh-xpmg-72qh
replace github.com/opencontainers/image-spec => github.com/opencontainers/image-spec v1.0.2
