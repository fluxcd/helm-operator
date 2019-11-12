package v2

import (
	"github.com/pkg/errors"
	"google.golang.org/grpc/status"

	"k8s.io/helm/pkg/chartutil"
	helmv2 "k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/proto/hapi/release"

	"github.com/fluxcd/helm-operator/pkg/helm"
)

type releaseResponse interface {
	GetRelease() *release.Release
}

func (h *HelmV2) UpgradeFromPath(chartPath string, releaseName string, values []byte,
	opts helm.UpgradeOptions) (*helm.Release, error) {
	// Load the chart from the given path
	chartRequested, err := chartutil.Load(chartPath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load chart from path [%s] for release [%s]", chartPath, releaseName)
	}

	// Validate the configured options
	if err := opts.Validate([]string{"atomic", "maxHistory"}); err != nil {
		h.logger.Log("warning", err.Error())
	}

	var res releaseResponse
	if opts.Install {
		res, err = h.client.InstallReleaseFromChart(
			chartRequested,
			opts.Namespace,
			helmv2.ReleaseName(releaseName),
			helmv2.ValueOverrides(values),
			helmv2.InstallDisableHooks(opts.DisableHooks),
			helmv2.InstallDryRun(opts.DryRun),
			helmv2.InstallWait(opts.Wait),
			helmv2.InstallTimeout(int64(opts.Timeout.Seconds())),
		)
	} else {
		res, err = h.client.UpdateReleaseFromChart(
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
	}
	if err != nil {
		if s, ok := status.FromError(err); ok {
			return nil, errors.New(s.Message())
		}
		return nil, err
	}
	return releaseToGenericRelease(res.GetRelease()), nil
}
