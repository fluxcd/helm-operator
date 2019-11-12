package v3

import (
	"github.com/pkg/errors"

	"helm.sh/helm/pkg/action"
	"helm.sh/helm/pkg/storage/driver"

	"github.com/fluxcd/helm-operator/pkg/helm"
)

func (h *HelmV3) Status(releaseName string, opts helm.StatusOptions) (*helm.Release, error) {
	cfg, cleanup, err := initActionConfig(h.kc, HelmOptions{Namespace: opts.Namespace})
	defer cleanup()
	if err != nil {
		return nil, errors.Wrap(err, "failed to setup Helm client")
	}

	client := action.NewStatus(cfg)
	client.Version = opts.Version

	rls, err := client.Run(releaseName)
	if err != nil {
		if err == driver.ErrReleaseNotFound {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "failed to retrieve status for release [%s]", releaseName)
	}
	return releaseToGenericRelease(rls), nil
}
