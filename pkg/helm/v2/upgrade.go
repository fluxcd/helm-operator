package v2

import (
	"github.com/pkg/errors"

	"k8s.io/helm/pkg/chartutil"
	helmv2 "k8s.io/helm/pkg/helm"

	"github.com/fluxcd/helm-operator/pkg/helm"
)

func (h *HelmV2) UpgradeFromPath(chartPath string, releaseName string, values []byte,
	opts helm.UpgradeOptions) (helm.Release, error) {

	// Load the chart from the given path
	chartRequested, err := chartutil.Load(chartPath)
	if err != nil {
		return helm.Release{}, errors.Wrapf(err, "failed to load chart from path [%s] for release [%s]", chartPath, releaseName)
	}

	// Validate the configured options
	if err := opts.Validate([]string{"atomic", "maxHistory"}); err != nil {
		h.logger.Log("warning", err.Error())
	}

	// Run upgrade
	res, err := h.client.UpdateReleaseFromChart(
		releaseName,
		chartRequested,
		helmv2.UpdateValueOverrides(values),
		helmv2.UpgradeDisableHooks(opts.DisableHooks),
		helmv2.UpgradeDryRun(opts.DryRun),
		helmv2.UpgradeForce(opts.Force),
		helmv2.UpgradeRecreate(opts.Recreate),
		helmv2.ReuseValues(opts.ReuseValues),
		helmv2.ResetValues(opts.ResetValues),
		helmv2.UpgradeTimeout(int64(opts.Timeout.Seconds())),
		helmv2.UpgradeWait(opts.Wait),
	)
	if err != nil {
		return helm.Release{}, errors.Wrapf(err, "failed to install chart release [%s]", releaseName)
	}
	return releaseToGenericRelease(res.Release), err
}
