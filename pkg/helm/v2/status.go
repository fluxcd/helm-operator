package v2

import (
	"strings"

	"github.com/pkg/errors"
	"google.golang.org/grpc/status"

	helmv2 "k8s.io/helm/pkg/helm"

	"github.com/fluxcd/helm-operator/pkg/helm"
)

func (h *HelmV2) Status(releaseName string, opts helm.StatusOptions) (*helm.Release, error) {
	// We use `ReleaseContent` here as `ReleaseStatus` does not return
	// the full release, which is required to construct a `helm.Release`.
	res, err := h.client.ReleaseContent(releaseName, helmv2.ContentReleaseVersion(int32(opts.Version)))
	if err != nil {
		if s, ok := status.FromError(err); ok {
			if strings.Contains(s.Message(), "not found") {
				return nil, nil
			}
			return nil, errors.New(s.Message())
		}
		return nil, err
	}
	return releaseToGenericRelease(res.Release), nil
}
