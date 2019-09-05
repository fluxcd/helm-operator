package v3

import (
	"fmt"

	"github.com/go-kit/kit/log"

	"helm.sh/helm/pkg/downloader"
	"helm.sh/helm/pkg/helmpath"
)

func (h *HelmV3) DependencyUpdate(chartPath string) error {
	out := &logWriter{h.logger}
	man := &downloader.Manager{
		Out:              out,
		ChartPath:        chartPath,
		RepositoryConfig: helmpath.ConfigPath("repositories.yaml"),
		RepositoryCache:  helmpath.CachePath("repository"),
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
