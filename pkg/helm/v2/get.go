package v2

import (
	"strings"

	helmv2 "k8s.io/helm/pkg/helm"

	"github.com/fluxcd/helm-operator/pkg/helm"
)

func (h *HelmV2) Get(releaseName string, opts helm.GetOptions) (*helm.Release, error) {
	res, err := h.client.ReleaseContent(releaseName, helmv2.ContentReleaseVersion(int32(opts.Version)))
	if err != nil {
		err = statusMessageErr(err)
		if strings.Contains(err.Error(), "not found") {
			return nil, nil
		}
		return nil, err
	}
	return releaseToGenericRelease(res.Release), nil
}

func (h *HelmV2) Status(releaseName string, opts helm.StatusOptions) (helm.Status, error) {
	res, err := h.client.ReleaseStatus(releaseName, helmv2.StatusReleaseVersion(int32(opts.Version)))
	if err != nil {
		err = statusMessageErr(err)
		if strings.Contains(err.Error(), "not found") {
			return "", nil
		}
		return "", err
	}
	return lookUpGenericStatus(res.Info.Status.Code), nil
}
