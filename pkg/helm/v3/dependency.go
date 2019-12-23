package v3

import (
	"helm.sh/helm/v3/pkg/downloader"

	"github.com/fluxcd/helm-operator/pkg/helm"
)

func (h *HelmV3) DependencyUpdate(chartPath string) error {
	repositoryConfigLock.RLock()
	defer repositoryConfigLock.RUnlock()

	out := helm.NewLogWriter(h.logger)
	man := &downloader.Manager{
		Out:              out,
		ChartPath:        chartPath,
		RepositoryConfig: repositoryConfig,
		RepositoryCache:  repositoryCache,
		Getters:          getters,
	}
	return man.Update()
}
