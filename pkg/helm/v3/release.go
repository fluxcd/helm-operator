package v3

import (
	"helm.sh/helm/pkg/chart"
	"helm.sh/helm/pkg/release"

	"github.com/fluxcd/helm-operator/pkg/helm"
)

// releaseToGenericRelease transforms a v3 release structure
// into a generic `helm.Release`
func releaseToGenericRelease(r *release.Release) helm.Release {
	return helm.Release{
		Name:      r.Name,
		Namespace: r.Namespace,
		Chart:     chartToGenericChart(r.Chart),
		Info:      infoToGenericInfo(r.Info),
		Values:    r.Config,
		Manifest:  r.Manifest,
		Version:   r.Version,
	}
}

// chartToGenericChart transforms a v3 chart structure into
// a generic `helm.Chart`
func chartToGenericChart(c *chart.Chart) *helm.Chart {
	return &helm.Chart{
		Name:       c.Name(),
		Version:    formatVersion(c),
		AppVersion: c.AppVersion(),
	}
}

// infoToGenericInfo transforms a v3 info structure into
// a generic `helm.Info`
func infoToGenericInfo(i *release.Info) *helm.Info {
	return &helm.Info{
		LastDeployed: i.LastDeployed,
		Description:  i.Description,
		Status:       lookUpGenericStatus(i.Status),
	}
}

// formatVersion formats the chart version, while taking
// into account that the metadata may actually be missing
// due to unknown reasons.
// https://github.com/kubernetes/helm/issues/1347
func formatVersion(c *chart.Chart) string {
	if c.Metadata == nil {
		return ""
	}
	return c.Metadata.Version
}

// lookUpGenericStatus looks up the generic status for the
// given v3 status
func lookUpGenericStatus(s release.Status) helm.Status {
	var statuses = map[release.Status]helm.Status{
		release.StatusUnknown:         helm.StatusUnknown,
		release.StatusDeployed:        helm.StatusDeployed,
		release.StatusUninstalled:     helm.StatusUninstalled,
		release.StatusSuperseded:      helm.StatusSuperseded,
		release.StatusFailed:          helm.StatusFailed,
		release.StatusUninstalling:    helm.StatusUninstalling,
		release.StatusPendingInstall:  helm.StatusPendingInstall,
		release.StatusPendingUpgrade:  helm.StatusPendingUpgrade,
		release.StatusPendingRollback: helm.StatusPendingRollback,
	}
	if status, ok := statuses[s]; ok {
		return status
	}
	return helm.StatusUnknown
}
