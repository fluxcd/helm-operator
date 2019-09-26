package v2

import (
	"github.com/pkg/errors"

	helmv2 "k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/proto/hapi/release"

	"github.com/fluxcd/helm-operator/pkg/helm"
)

func (h *HelmV2) History(releaseName string, opts helm.HistoryOptions) ([]helm.Release, error) {
	res, err := h.client.ReleaseHistory(releaseName, helmv2.WithMaxHistory(int32(opts.Max)))
	if err != nil {
		return make([]helm.Release, 0), errors.Wrapf(err, "failed to retrieve history for [%s]", releaseName)
	}
	return getReleaseHistory(res.Releases), nil
}

func getReleaseHistory(rls []*release.Release) []helm.Release {
	history := make([]helm.Release, len(rls))
	for i := len(rls) - 1; i >= 0; i-- {
		r := rls[i]
		history = append(history, releaseToGenericRelease(r))
	}
	return history
}
