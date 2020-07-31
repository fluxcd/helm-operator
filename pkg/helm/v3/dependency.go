package v3

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"helm.sh/helm/v3/pkg/downloader"

	"github.com/fluxcd/helm-operator/pkg/utils"
)

func (h *HelmV3) DependencyUpdate(chartPath string) error {
	// Garbage collect before the dependency update so that
	// anonymous files from previous runs are cleared, with
	// a safe guard time offset to not touch any files in
	// use.
	garbageCollect(repositoryCache, time.Second * 300)
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

// garbageCollect walks over the files in the given path and deletes
// any anonymous index file with a mod time older than the given
// duration.
func garbageCollect(path string, olderThan time.Duration) {
	now := time.Now()
	filepath.Walk(path, func(p string, f os.FileInfo, err error) error {
		if err != nil || f.IsDir() {
			return nil
		}
		if strings.HasSuffix(f.Name(), "=-index.yaml") || strings.HasSuffix(f.Name(), "=-charts.txt") {
			if now.Sub(f.ModTime()) > olderThan {
				return os.Remove(p)
			}
		}
		return nil
	})
}
