package v3

import (
	"github.com/fluxcd/helm-operator/pkg/helm"
	"github.com/pkg/errors"

	"helm.sh/helm/pkg/action"
)

func (h *HelmV3) Uninstall(releaseName string, opts helm.UninstallOptions) error {
	cfg, cleanup, err := initActionConfig(h.kc, HelmOptions{Namespace: opts.Namespace})
	defer cleanup()
	if err != nil {
		return errors.Wrap(err, "failed to setup Helm client")
	}

	client := action.NewUninstall(cfg)

	// Set all configured options
	client.DisableHooks = opts.DisableHooks
	client.DryRun = opts.DryRun
	client.KeepHistory = opts.KeepHistory
	client.Timeout = opts.Timeout

	// Run uninstall
	if _, err := client.Run(releaseName); err != nil {
		return errors.Wrapf(err, "failed to uninstall release [%s]", releaseName)
	}
	return nil
}
