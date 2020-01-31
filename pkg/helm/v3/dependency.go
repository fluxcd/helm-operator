package v3

import (
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"

	"github.com/fluxcd/helm-operator/pkg/helm"
)

func (h *HelmV3) DependencyUpdate(chartPath string) error {
	out := helm.NewLogWriter(h.logger)
	man := &downloader.Manager{
		Out:              out,
		ChartPath:        chartPath,
		RepositoryConfig: repositoryConfig,
		RepositoryCache:  repositoryCache,
		Getters: getter.All(&cli.EnvSettings{
			RepositoryConfig: repositoryConfig,
			RepositoryCache:  repositoryCache,
			PluginsDirectory: pluginsDir,
		}),
	}
	return man.Update()
}
