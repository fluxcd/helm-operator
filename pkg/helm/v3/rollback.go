package v3

import (
	"github.com/pkg/errors"

	"helm.sh/helm/pkg/action"

	"github.com/fluxcd/helm-operator/pkg/helm"
)

func (h *HelmV3) Rollback(releaseName string, opts helm.RollbackOptions) (helm.Release, error) {
	cfg, cleanup, err := initActionConfig(h.kc, HelmOptions{Namespace: opts.Namespace})
	defer cleanup()
	if err != nil {
		return helm.Release{}, errors.Wrap(err, "failed to setup Helm client")
	}

	client := action.NewRollback(cfg)

	// Set all configured options
	client.Version = opts.Version
	client.Timeout = opts.Timeout
	client.Wait = opts.Wait
	client.DisableHooks = opts.DisableHooks
	client.DryRun = opts.DryRun
	client.Recreate = opts.Recreate
	client.Force = opts.Force

	// Run rollback
	err = client.Run(releaseName)
	if err != nil {
		return helm.Release{}, errors.Wrapf(err, "failed to perform rollback for release [%s]", releaseName)
	}
	// As rolling back does no longer return information about
	// the release in v3 we need to make an additional call to
	// get the release we rolled back to.
	return h.Status(releaseName, helm.StatusOptions{})
}
