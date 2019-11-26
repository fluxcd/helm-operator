package v3

import (
	"fmt"

	"github.com/go-kit/kit/log"

	"helm.sh/helm/v3/pkg/downloader"
)

func (h *HelmV3) DependencyUpdate(chartPath string) error {
	repositoryConfigLock.RLock()
	defer repositoryConfigLock.RUnlock()

	out := &logWriter{h.logger}
	man := &downloader.Manager{
		Out:              out,
		ChartPath:        chartPath,
		RepositoryConfig: repositoryConfig,
		RepositoryCache:  repositoryCache,
	}
	return man.Update()
}

// logWriter wraps a `log.Logger` so it can be used as an `io.Writer`
type logWriter struct {
	log.Logger
}

func (l *logWriter) Write(p []byte) (n int, err error) {
	origLen := len(p)
	if len(p) > 0 && p[len(p)-1] == '\n' {
		p = p[:len(p)-1] // Cut terminating newline
	}
	l.Log("info", fmt.Sprintf("%s", p))
	return origLen, nil
}
