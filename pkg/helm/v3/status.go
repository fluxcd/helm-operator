package v3

import (
	"github.com/pkg/errors"

	"helm.sh/helm/pkg/action"

	"github.com/fluxcd/helm-operator/pkg/helm"
)

func (h *HelmV3) Status(releaseName string, opts helm.StatusOptions) (helm.Release, error) {
	cfg, cleanup, err := initActionConfig(h.kc, HelmOptions{Namespace: opts.Namespace})
	defer cleanup()
	if err != nil {
		return helm.Release{}, errors.Wrap(err, "failed to setup Helm client")
	}

	client := action.NewStatus(cfg)
	client.Version = opts.Version

	rls, err := client.Run(releaseName)
	if err != nil {
		return helm.Release{}, errors.Wrapf(err, "failed to retrieve status for release [%s]", releaseName)
	}
	return releaseToGenericRelease(rls), nil
}
