module github.com/fluxcd/helm-operator

go 1.13

require (
	github.com/bugsnag/bugsnag-go v1.5.3 // indirect
	github.com/fluxcd/flux v1.15.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-kit/kit v0.9.0
	github.com/google/go-cmp v0.3.1
	github.com/gorilla/handlers v1.4.2 // indirect
	github.com/gorilla/mux v1.7.1
	github.com/gosuri/uitable v0.0.3 // indirect
	github.com/instrumenta/kubeval v0.0.0-20190804145309-805845b47dfc
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v1.2.1
	github.com/shurcooL/httpfs v0.0.0-20190707220628-8d4bc4ba7749
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.4.0
	github.com/yvasiyarov/gorelic v0.0.7 // indirect
	google.golang.org/grpc v1.24.0
	helm.sh/helm/v3 v3.0.1
	k8s.io/api v0.0.0-20191016110408-35e52d86657a // kubernetes-1.16.2
	k8s.io/apiextensions-apiserver v0.0.0-20191016113550-5357c4baaf65 // kubernetes-1.16.2
	k8s.io/apimachinery v0.0.0-20191004115801-a2eda9f80ab8 // kubernetes-1.16.2
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/helm v2.14.3+incompatible
	k8s.io/klog v1.0.0
	k8s.io/kubectl v0.0.0-20191016120415-2ed914427d51 // kubernetes-1.16.2
	rsc.io/letsencrypt v0.0.3 // indirect
)

// Transitive requirement from Flux: remove when https://github.com/docker/distribution/pull/2905 is released.
replace github.com/docker/distribution => github.com/2opremio/distribution v0.0.0-20190419185413-6c9727e5e5de

// Transitive requirement from Helm.
replace github.com/docker/docker => github.com/docker/docker v0.7.3-0.20190327010347-be7ac8be2ae0

// Force upgrade because of a transitive downgrade.
// github.com/fluxcd/helm-operator
// +-> github.com/fluxcd/flux@v1.15.0
//     +-> k8s.io/client-go@v11.0.0+incompatible
//     +-> github.com/fluxcd/helm-operator@v1.0.0-rc1
//         +-> k8s.io/client-go@v11.0.0+incompatible
//         +-> github.com/weaveworks/flux@v0.0.0-20190729133003-c78ccd3706b5
//             +-> k8s.io/client-go@v11.0.0+incompatible
replace k8s.io/client-go => k8s.io/client-go v0.0.0-20191016111102-bec269661e48 // kubernetes-1.16.2
