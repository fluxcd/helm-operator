package v3

import (
	"github.com/pkg/errors"

	"helm.sh/helm/pkg/action"
	"helm.sh/helm/pkg/chart/loader"
	"helm.sh/helm/pkg/chartutil"

	"github.com/fluxcd/helm-operator/pkg/helm"
)

func (h *HelmV3) UpgradeFromPath(chartPath string, releaseName string, values []byte,
	opts helm.UpgradeOptions) (helm.Release, error) {

	cfg, cleanup, err := initActionConfig(h.kc, HelmOptions{Namespace: opts.Namespace})
	defer cleanup()
	if err != nil {
		return helm.Release{}, errors.Wrap(err, "failed to setup Helm client")
	}

	client := action.NewUpgrade(cfg)
	client.Namespace = opts.Namespace

	// Load the chart from the given path, this also ensures that
	// all chart dependencies are present
	chartRequested, err := loader.Load(chartPath)
	if err != nil {
		return helm.Release{}, errors.Wrapf(err, "failed to load chart from path [%s] for release [%s]", chartPath, releaseName)
	}

	// Read and set values
	val, err := chartutil.ReadValues(values)
	if err != nil {
		return helm.Release{}, errors.Wrap(err, "failed to read values")
	}

	// Validate the configured options
	if err := opts.Validate([]string{}); err != nil {
		h.logger.Log("warning", err.Error())
	}

	// Set all configured options
	client.Atomic = opts.Atomic
	client.DisableHooks = opts.DisableHooks
	client.DryRun = opts.DryRun
	client.Force = opts.Force
	client.MaxHistory = opts.MaxHistory
	client.ResetValues = opts.ResetValues
	client.ReuseValues = opts.ReuseValues
	client.Timeout = opts.Timeout
	client.Wait = opts.Wait

	// Run upgrade
	res, err := client.Run(releaseName, chartRequested, val.AsMap())
	if err != nil {
		return helm.Release{}, errors.Wrapf(err, "failed to upgrade chart for release [%s]", releaseName)
	}
	return releaseToGenericRelease(res), err
}
