package v2

import (
	"fmt"

	"k8s.io/helm/pkg/chartutil"
)

func (h *HelmV2) GetChartRevision(chartPath string) (string, error) {
	chartRequested, err := chartutil.Load(chartPath)
	if err != nil {
		return "", fmt.Errorf("failed to load chart to determine revision: %w", err)
	}
	return chartRequested.Metadata.Version, nil
}
