package v3

import (
	"strings"

	"github.com/helm/helm-2to3/pkg/common"
	helm2 "github.com/helm/helm-2to3/pkg/v2"
	helm3 "github.com/helm/helm-2to3/pkg/v3"
)

// Converter Converts a given helm 2 release with all its release versions to helm 3 format and deletes the old release from tiller
type Converter struct {
	TillerNamespace  string
	KubeConfig       string // file path to kubeconfig
	TillerOutCluster bool
	StorageType      string
}

// V2ReleaseExists helps you check if a helm v2 release exists or not
func (c Converter) V2ReleaseExists(releaseName string) (bool, error) {
	retrieveOpts := helm2.RetrieveOptions{
		ReleaseName:      releaseName,
		TillerNamespace:  c.TillerNamespace,
		TillerOutCluster: c.TillerOutCluster,
		StorageType:      c.StorageType,
	}
	kubeConfig := common.KubeConfig{
		File: c.KubeConfig,
	}
	v2Releases, err := helm2.GetReleaseVersions(retrieveOpts, kubeConfig)

	// We check for the error message content because
	// Helm 2to3 returns an error if it doesn't find release versions
	if err != nil && !strings.Contains(err.Error(), "has no deployed releases") {
		return false, err
	}
	return len(v2Releases) > 0, nil
}

// Convert attempts to convert the given release name from v2 to v3.
func (c Converter) Convert(releaseName string, dryRun bool) error {
	retrieveOpts := helm2.RetrieveOptions{
		ReleaseName:     releaseName,
		TillerNamespace: c.TillerNamespace,
	}
	kubeConfig := common.KubeConfig{
		File: c.KubeConfig,
	}
	v2Releases, err := helm2.GetReleaseVersions(retrieveOpts, kubeConfig)
	if err != nil {
		return err
	}

	if !dryRun {
		for _, v2Release := range v2Releases {
			v3Release, err := helm3.CreateRelease(v2Release)
			if err != nil {
				return err
			}
			if err := helm3.StoreRelease(v3Release, kubeConfig); err != nil {
				return err
			}
		}
	}

	if err := helm2.DeleteAllReleaseVersions(retrieveOpts, kubeConfig, dryRun); err != nil {
		return err
	}
	return nil
}
