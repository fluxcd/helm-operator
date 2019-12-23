package v2

import (
	"k8s.io/helm/pkg/downloader"
	"k8s.io/helm/pkg/repo"
	"k8s.io/helm/pkg/urlutil"

	"github.com/fluxcd/helm-operator/pkg/helm"
)

func (h *HelmV2) Pull(ref, version, dest string) (string, error) {
	repositoryConfigLock.RLock()
	defer repositoryConfigLock.RUnlock()

	out := helm.NewLogWriter(h.logger)
	c := downloader.ChartDownloader{
		Out:      out,
		HelmHome: helmHome(),
		Verify:   downloader.VerifyNever,
		Getters:  getters,
	}
	d, _, err := c.DownloadTo(ref, version, dest)
	return d, err
}

func (h *HelmV2) PullWithRepoURL(repoURL, name, version, dest string) (string, error) {
	// This resolves the repo URL, chart name and chart version to a
	// URL for the chart. To be able to resolve the chart name and
	// version to a URL, we have to have the index file; and to have
	// that, we may need to authenticate. The credentials will be in
	// the repository config.
	repositoryConfigLock.RLock()
	repoFile, err := loadRepositoryConfig()
	repositoryConfigLock.RUnlock()
	if err != nil {
		return "", err
	}

	// Now find the entry for the repository, if there is one. If not,
	// we'll assume there's no auth needed.
	repoEntry := &repo.Entry{}
	repoEntry.URL = repoURL
	for _, entry := range repoFile.Repositories {
		if urlutil.Equal(repoEntry.URL, entry.URL) {
			repoEntry = entry
			// Ensure we have the repository index as this is
			// later used by Helm.
			if r, err := repo.NewChartRepository(repoEntry, getters); err == nil {
				r.DownloadIndexFile(repositoryCache)
			}
			break
		}
	}

	// Look up the full URL of the chart with the collected credentials
	// and given chart name and version.
	chartURL, err := repo.FindChartInAuthRepoURL(repoEntry.URL, repoEntry.Username, repoEntry.Password, name, version,
		repoEntry.CertFile, repoEntry.KeyFile, repoEntry.CAFile, getters)
	if err != nil {
		return "", err
	}

	return h.Pull(chartURL, version, dest)
}
