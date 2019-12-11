package v2

import (
	"fmt"

	"github.com/go-kit/kit/log"

	"k8s.io/helm/pkg/downloader"
)

func (h *HelmV2) DependencyUpdate(chartPath string) error {
	repositoryConfigLock.RLock()
	defer repositoryConfigLock.RUnlock()

	out := &logWriter{h.logger}
	man := downloader.Manager{
		Out: out,
		ChartPath: chartPath,
		HelmHome: helmHome(),
		Getters: getters,
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
