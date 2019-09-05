package v2

import (
	"github.com/pkg/errors"

	helmv2 "k8s.io/helm/pkg/helm"

	"github.com/fluxcd/helm-operator/pkg/helm"
)

func (h *HelmV2) Status(releaseName string, opts helm.StatusOptions) (helm.Release, error) {
	// We use `ReleaseContent` here as `ReleaseStatus` does not return
	// the full release, which is required to construct a `helm.Release`.
	res, err := h.client.ReleaseContent(releaseName, helmv2.ContentReleaseVersion(int32(opts.Version)))
	if err != nil {
		return helm.Release{}, errors.Wrapf(err, "failed to retrieve status for [%s]", releaseName)
	}
	return releaseToGenericRelease(res.Release), nil
}
