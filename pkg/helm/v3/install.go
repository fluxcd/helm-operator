package v3

import (
	"github.com/pkg/errors"
	"helm.sh/helm/pkg/action"
	"helm.sh/helm/pkg/chart/loader"
	"helm.sh/helm/pkg/chartutil"

	"github.com/fluxcd/helm-operator/pkg/helm"
)

func (h *HelmV3) InstallFromPath(chartPath string, releaseName string, values []byte,
	opts helm.InstallOptions) (helm.Release, error) {

	cfg, cleanup, err := initActionConfig(h.kc, HelmOptions{Namespace: opts.Namespace})
	defer cleanup()
	if err != nil {
		return helm.Release{}, errors.Wrap(err, "failed to setup Helm client")
	}

	client := action.NewInstall(cfg)
	client.Namespace = opts.Namespace
	client.ReleaseName = releaseName

	// Load the chart from the given path, this also ensures that
	// all chart dependencies are present
	chartRequested, err := loader.Load(chartPath)
	if err != nil {
		return helm.Release{}, errors.Wrapf(err, "failed to load chart from path [%s]", chartPath)
	}

	// Read and set values
	val, err := chartutil.ReadValues(values)
	if err != nil {
		return helm.Release{}, errors.Wrap(err, "failed to read values")
	}

	// Validate the configured options
	if err := opts.Validate([]string{"disableCRDHooks"}); err != nil {
		h.logger.Log("warning", err.Error())
	}

	// Set all configured options
	client.Atomic = opts.Atomic
	client.ClientOnly = opts.ClientOnly
	client.DependencyUpdate = opts.DependencyUpdate
	client.DisableHooks = opts.DisableHooks
	client.DryRun = opts.DryRun
	client.Replace = opts.Replace
	client.Timeout = opts.Timeout
	client.Wait = opts.Wait

	// Run installation
	rel, err := client.Run(chartRequested, val.AsMap())
	if err != nil {
		return helm.Release{}, errors.Wrapf(err, "failed to install chart release [%s]", releaseName)
	}
	return releaseToGenericRelease(rel), err
}
