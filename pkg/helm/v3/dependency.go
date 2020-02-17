package v3

import (
	"helm.sh/helm/v3/pkg/downloader"

	"github.com/fluxcd/helm-operator/pkg/utils"
)

func (h *HelmV3) DependencyUpdate(chartPath string) error {
	out := utils.NewLogWriter(h.logger)
	man := &downloader.Manager{
		Out:              out,
		ChartPath:        chartPath,
		RepositoryConfig: repositoryConfig,
		RepositoryCache:  repositoryCache,
		Getters:          getterProviders(),
	}
	return man.Update()
}
