package v3

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/go-kit/kit/log"

	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/klog"
	"k8s.io/kubectl/pkg/cmd/util"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/helmpath"
	"helm.sh/helm/v3/pkg/kube"
	"helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"

	"github.com/fluxcd/helm-operator/pkg/helm"
)

const VERSION = "v3"

var (
	defaultClusterName = "in-cluster"
	repositoryConfig   = helmpath.ConfigPath("repositories.yaml")
	repositoryCache    = helmpath.CachePath("repository")
)

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

func (h *HelmV3) Version() string {
	return VERSION
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
		Factory: util.NewFactory(cfgFlags),
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

	caData, err := ioutil.ReadFile(kc.CAFile)
	if err != nil {
		return "", "", cleanup, err
	}

	c := newConfig(kc.Host, kc.Username, kc.BearerToken, caData)
	tmpFullPath := tmpDir + "/config"
	if err := clientcmd.WriteToFile(c, tmpFullPath); err != nil {
		return "", "", cleanup, err
	}

	return tmpFullPath, c.CurrentContext, cleanup, nil
}

func newConfig(host, username, token string, caCert []byte) clientcmdapi.Config {

	contextName := fmt.Sprintf("%s@%s", username, defaultClusterName)

	return clientcmdapi.Config{
		Clusters: map[string]*clientcmdapi.Cluster{
			defaultClusterName: {
				Server:                   host,
				CertificateAuthorityData: caCert,
			},
		},
		Contexts: map[string]*clientcmdapi.Context{
			contextName: {
				Cluster:  defaultClusterName,
				AuthInfo: username,
			},
		},
		AuthInfos: map[string]*clientcmdapi.AuthInfo{
			username: {
				Token: token,
			},
		},
		CurrentContext: contextName,
	}
}
