package v2

import (
	"github.com/pkg/errors"
	"k8s.io/helm/pkg/chartutil"
	helmv2 "k8s.io/helm/pkg/helm"

	"github.com/fluxcd/helm-operator/pkg/helm"
)

func (h *HelmV2) InstallFromPath(chartPath string, releaseName string, values []byte,
	opts helm.InstallOptions) (helm.Release, error) {

	// Load the chart from the given path
	chartRequested, err := chartutil.Load(chartPath)
	if err != nil {
		return helm.Release{}, errors.Wrapf(err, "failed to load chart from path [%s]", chartPath)
	}

	// Validate the configured options
	if err := opts.Validate([]string{"clientOnly", "dependencyUpdate", "atomic"}); err != nil {
		h.logger.Log("warning", err.Error())
	}

	// Run installation
	res, err := h.client.InstallReleaseFromChart(
		chartRequested,
		opts.Namespace,
		helmv2.ReleaseName(releaseName),
		helmv2.ValueOverrides(values),
		helmv2.InstallDisableHooks(opts.DisableHooks),
		helmv2.InstallDisableCRDHook(opts.DisableCRDHooks),
		helmv2.InstallReuseName(opts.Replace),
		helmv2.InstallDryRun(opts.DryRun),
		helmv2.InstallWait(opts.Wait),
		helmv2.InstallTimeout(int64(opts.Timeout.Seconds())),
	)
	if err != nil {
		return helm.Release{}, errors.Wrap(err, "failed to install chartRequested release")
	}
	return releaseToGenericRelease(res.Release), err
}
