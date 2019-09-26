package v3

import (
	"github.com/pkg/errors"

	"helm.sh/helm/pkg/action"
	"helm.sh/helm/pkg/release"
	"helm.sh/helm/pkg/releaseutil"

	"github.com/fluxcd/helm-operator/pkg/helm"
)

func (h *HelmV3) History(releaseName string, opts helm.HistoryOptions) ([]helm.Release, error) {
	cfg, cleanup, err := initActionConfig(h.kc, HelmOptions{Namespace: opts.Namespace})
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

	var rels []*release.Release
	for i := 0; i < min(len(hist), client.Max); i++ {
		rels = append(rels, hist[i])
	}

	if len(rels) == 0 {
		return make([]helm.Release, 0), nil
	}

	return getReleaseHistory(hist), nil
}

func getReleaseHistory(rls []*release.Release) []helm.Release {
	history := make([]helm.Release, len(rls))
	for i := len(rls) - 1; i >= 0; i-- {
		r := rls[i]
		history = append(history, releaseToGenericRelease(r))
	}
	return history
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
