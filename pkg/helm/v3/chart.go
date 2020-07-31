package v3

import (
	"fmt"

	"helm.sh/helm/v3/pkg/chart/loader"
)

func (h *HelmV3) GetChartRevision(chartPath string) (string, error) {
	chartRequested, err := loader.Load(chartPath)
	if err != nil {
		return "", fmt.Errorf("failed to load chart to determine revision: %w", err)
	}
	return chartRequested.Metadata.Version, nil
}
