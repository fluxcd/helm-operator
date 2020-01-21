package v3

import (
	"github.com/pkg/errors"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/storage/driver"

	"github.com/fluxcd/helm-operator/pkg/helm"
)

func (h *HelmV3) Get(releaseName string, opts helm.GetOptions) (*helm.Release, error) {
	cfg, cleanup, err := h.initActionConfig(HelmOptions{Namespace: opts.Namespace})
	defer cleanup()
	if err != nil {
		return nil, errors.Wrap(err, "failed to setup Helm client")
	}

	client := action.NewGet(cfg)
	client.Version = opts.Version

	res, err := client.Run(releaseName)
	if err != nil {
		if err == driver.ErrReleaseNotFound {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "failed to retrieve release [%s]", releaseName)
	}
	return releaseToGenericRelease(res), err
}
