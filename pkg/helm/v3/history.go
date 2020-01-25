package v3

import (
	"github.com/pkg/errors"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/releaseutil"

	"github.com/fluxcd/helm-operator/pkg/helm"
)

func (h *HelmV3) History(releaseName string, opts helm.HistoryOptions) ([]*helm.Release, error) {
	cfg, cleanup, err := h.initActionConfig(HelmOptions{Namespace: opts.Namespace})
	defer cleanup()

	if err != nil {
		return nil, errors.Wrap(err, "failed to setup Helm client")
	}

	client := action.NewHistory(cfg)
	client.Max = opts.Max

	hist, err := client.Run(releaseName)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to retrieve history for [%s]", releaseName)
	}

	releaseutil.Reverse(hist, releaseutil.SortByRevision)

	var rels []*helm.Release
	for i := 0; i < min(len(hist), client.Max); i++ {
		rels = append(rels, releaseToGenericRelease(hist[i]))
	}
	return rels, nil
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
