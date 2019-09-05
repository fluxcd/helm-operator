package v3

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/go-kit/kit/log"

	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/kubeconfig"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"

	"helm.sh/helm/pkg/action"
	"helm.sh/helm/pkg/kube"
	"helm.sh/helm/pkg/storage"
	"helm.sh/helm/pkg/storage/driver"

	"github.com/fluxcd/helm-operator/pkg/helm"
)

const VERSION = "v3"

var defaultClusterName = "in-cluster"

type HelmOptions struct {
	Driver    string
	Namespace string
}

type HelmV3 struct {
	kc     *rest.Config
	logger log.Logger
}

// New creates a new HelmV3 client
func New(logger log.Logger, kubeConfig *rest.Config) helm.Client {
	// Add CRDs to the scheme. They are missing by default but required
	// by Helm v3.
	if err := apiextv1beta1.AddToScheme(scheme.Scheme); err != nil {
		// This should never happen.
		panic(err)
	}
	return &HelmV3{
		kc:     kubeConfig,
		logger: logger,
	}
}

// initActionConfig initializes the configuration for the action,
// like the namespace it should be executed in and the storage driver.
func initActionConfig(kubeConfig *rest.Config, opts HelmOptions) (*action.Configuration, func(), error) {
	path, ctx, cleanup, err := writeTempKubeConfig(kubeConfig)
	if err != nil {
		return nil, nil, err
	}

	cfgFlags := kube.GetConfig(path, ctx, opts.Namespace)

	// We simply construct the client here instead of using `kube.New`
	// to prevent concurrency issues due to the `AddToScheme` call it
	// makes.
	kc := &kube.Client{
		Factory: cmdutil.NewFactory(cfgFlags),
		Log:     klog.Infof,
	}

	clientset, err := kc.Factory.KubernetesClientSet()
	if err != nil {
		return nil, cleanup, err
	}

	namespace := opts.Namespace

	var store *storage.Storage
	switch opts.Driver {
	case "secret", "secrets", "":
		d := driver.NewSecrets(clientset.CoreV1().Secrets(namespace))
		d.Log = klog.Infof
		store = storage.Init(d)
	case "configmap", "configmaps":
		d := driver.NewConfigMaps(clientset.CoreV1().ConfigMaps(namespace))
		d.Log = klog.Infof
		store = storage.Init(d)
	case "memory":
		d := driver.NewMemory()
		store = storage.Init(d)
	default:
		return nil, cleanup, fmt.Errorf("unknown Client storage driver [%s]", opts.Driver)
	}

	return &action.Configuration{
		RESTClientGetter: cfgFlags,
		Releases:         store,
		KubeClient:       kc,
		Log:              klog.Infof,
	}, cleanup, nil
}

// writeTempKubeConfig writes the given Config to a temporary location
// to be used by Client as a `.kube/config` file. The reason we do this
// is to be able to utilize Kubernetes' in-cluster discovery. It returns
// a cleanup function.
func writeTempKubeConfig(kc *rest.Config) (string, string, func(), error) {
	tmpDir, err := ioutil.TempDir("", "helmv3")
	if err != nil {
		return "", "", func() {}, err
	}
	cleanup := func() { os.RemoveAll(tmpDir) }

	token, err := ioutil.ReadFile(kc.CAFile)
	if err != nil {
		return "", "", cleanup, err
	}
	c := kubeconfig.CreateWithToken(kc.Host, defaultClusterName, kc.Username, token, kc.BearerToken)
	tmpFullPath := tmpDir + "/config"
	if err := kubeconfig.WriteToDisk(tmpFullPath, c); err != nil {
		return "", "", cleanup, err
	}
	return tmpFullPath, fmt.Sprintf("%s@%s", kc.Username, defaultClusterName), cleanup, nil
}
