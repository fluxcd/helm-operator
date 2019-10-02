module github.com/fluxcd/helm-operator

go 1.12

require (
	github.com/fluxcd/flux v1.15.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-kit/kit v0.9.0
	github.com/golang/protobuf v1.3.2
	github.com/google/go-cmp v0.3.0
	github.com/gorilla/mux v1.7.1
	github.com/instrumenta/kubeval v0.0.0-20190804145309-805845b47dfc
	github.com/ncabatoff/go-seq v0.0.0-20180805175032-b08ef85ed833
	github.com/prometheus/client_golang v1.1.0
	github.com/shurcooL/httpfs v0.0.0-20190707220628-8d4bc4ba7749
	github.com/spf13/pflag v1.0.3
	github.com/stretchr/testify v1.3.0
	github.com/weaveworks/flux v0.0.0-20190729133003-c78ccd3706b5
	k8s.io/api v0.0.0-20190313235455-40a48860b5ab
	k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/code-generator v0.0.0-20190511023357-639c964206c2
	k8s.io/gengo v0.0.0-20190907103519-ebc107f98eab // indirect
	k8s.io/helm v2.13.1+incompatible
	k8s.io/klog v0.3.3
)

// this is required to avoid
//    github.com/docker/distribution@v0.0.0-00010101000000-000000000000: unknown revision 000000000000
// because flux also replaces it, and we depend on flux
replace github.com/docker/distribution => github.com/2opremio/distribution v0.0.0-20190419185413-6c9727e5e5de

// The following pin these libs to `kubernetes-1.14.4` (by initially
// giving the version as that tag, and letting go mod fill in its idea of
// the version).
// The libs are thereby kept compatible with client-go v11, which is
// itself compatible with Kubernetes 1.14.

replace (
	k8s.io/api => k8s.io/api v0.0.0-20190708174958-539a33f6e817
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20190708180123-608cd7da68f7
	k8s.io/client-go => k8s.io/client-go v11.0.0+incompatible
	k8s.io/code-generator => k8s.io/code-generator v0.0.0-20190311093542-50b561225d70
	k8s.io/component-base => k8s.io/component-base v0.0.0-20190708175518-244289f83105
)
