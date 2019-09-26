module github.com/fluxcd/helm-operator

go 1.12

require (
	github.com/bugsnag/bugsnag-go v1.5.3 // indirect
	github.com/deislabs/oras v0.7.0 // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/go-kit/kit v0.9.0
	github.com/golang/protobuf v1.3.2 // indirect
	github.com/google/go-cmp v0.3.0
	github.com/googleapis/gnostic v0.3.0 // indirect
	github.com/gorilla/handlers v1.4.2 // indirect
	github.com/gorilla/mux v1.7.1
	github.com/gosuri/uitable v0.0.3 // indirect
	github.com/hashicorp/golang-lru v0.5.3 // indirect
	github.com/instrumenta/kubeval v0.0.0-20190804145309-805845b47dfc
	github.com/json-iterator/go v1.1.7 // indirect
	github.com/mattn/go-runewidth v0.0.4 // indirect
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v0.9.3-0.20190127221311-3c4408c8b829
	github.com/shurcooL/httpfs v0.0.0-20190707220628-8d4bc4ba7749
	github.com/spf13/pflag v1.0.3
	github.com/stretchr/testify v1.3.0
	github.com/weaveworks/flux v0.0.0-20190729133003-c78ccd3706b5
	github.com/yvasiyarov/gorelic v0.0.7 // indirect
	golang.org/x/crypto v0.0.0-20190701094942-4def268fd1a4 // indirect
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45 // indirect
	helm.sh/helm v3.0.0-beta.3+incompatible
	k8s.io/api v0.0.0
	k8s.io/apiextensions-apiserver v0.0.0
	k8s.io/apimachinery v0.0.0
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/code-generator v0.0.0
	k8s.io/helm v2.14.3+incompatible
	k8s.io/klog v0.3.3
	k8s.io/kubernetes v1.15.3
	k8s.io/utils v0.0.0-20190712204705-3dccf664f023 // indirect
	rsc.io/letsencrypt v0.0.3 // indirect
)

// this is required to avoid
//    github.com/docker/distribution@v0.0.0-00010101000000-000000000000: unknown revision 000000000000
// because flux also replaces it, and we depend on flux
replace (
	github.com/docker/distribution => github.com/2opremio/distribution v0.0.0-20190419185413-6c9727e5e5de
	github.com/docker/docker => github.com/docker/docker v0.7.3-0.20190327010347-be7ac8be2ae0
)

// The following pin these libs to `kubernetes-1.15.3` (by initially
// giving the version as that tag, and letting go mod fill in its idea of
// the version).
// The libs are thereby kept compatible with client-go v12, which is
// itself compatible with Kubernetes 1.15.
replace (
	k8s.io/api => k8s.io/api v0.0.0-20190819141258-3544db3b9e44
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190819143637-0dbe462fe92d
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190817020851-f2f3a405f61d
	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20190819142446-92cc630367d0
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20190819144027-541433d7ce35
	k8s.io/client-go => k8s.io/client-go v0.0.0-20190819141724-e14f31a72a77
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20190819145148-d91c85d212d5
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.0.0-20190819145008-029dd04813af
	k8s.io/code-generator => k8s.io/code-generator v0.0.0-20190612205613-18da4a14b22b
	k8s.io/component-base => k8s.io/component-base v0.0.0-20190819141909-f0f7c184477d
	k8s.io/cri-api => k8s.io/cri-api v0.0.0-20190817025403-3ae76f584e79
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.0.0-20190819145328-4831a4ced492
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20190819142756-13daafd3604f
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.0.0-20190819144832-f53437941eef
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.0.0-20190819144346-2e47de1df0f0
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.0.0-20190819144657-d1a724e0828e
	k8s.io/kubectl => k8s.io/kubectl v0.0.0-20190602132728-7075c07e78bf
	k8s.io/kubelet => k8s.io/kubelet v0.0.0-20190819144524-827174bad5e8
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.0.0-20190819145509-592c9a46fd00
	k8s.io/metrics => k8s.io/metrics v0.0.0-20190819143841-305e1cef1ab1
	k8s.io/node-api => k8s.io/node-api v0.0.0-20190819145652-b61681edbd0a
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.0.0-20190819143045-c84c31c165c4
	k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.0.0-20190819144209-f9ca4b649af0
	k8s.io/sample-controller => k8s.io/sample-controller v0.0.0-20190819143301-7c475f5e1313
)
