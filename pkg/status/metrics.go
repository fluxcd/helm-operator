package status

import (
	v1 "github.com/fluxcd/helm-operator/pkg/apis/helm.fluxcd.io/v1"
	"github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

const (
	LabelNamespace   = "namespace"
	LabelReleaseName = "release_name"
)

var phaseToGaugeValue = map[v1.HelmReleasePhase]float64{
	// Unknown is mapped to 0
	v1.HelmReleasePhaseChartFetchFailed: -4,
	v1.HelmReleasePhaseFailed:           -3,
	v1.HelmReleasePhaseRollbackFailed:   -2,
	v1.HelmReleasePhaseRolledBack:       -1,
	v1.HelmReleasePhaseRollingBack:      1,
	v1.HelmReleasePhaseInstalling:       2,
	v1.HelmReleasePhaseUpgrading:        3,
	v1.HelmReleasePhaseChartFetched:     4,
	v1.HelmReleasePhaseSucceeded:        5,
}

var (
	releasePhase = prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
		Namespace: "flux",
		Subsystem: "helm_operator",
		Name:      "release_phase_info",
		Help:      "Current HelmRelease phase.",
	}, []string{LabelNamespace, LabelReleaseName})
)

func SetReleasePhaseGauge(phase v1.HelmReleasePhase, namespace, releaseName string) {
	value, ok := phaseToGaugeValue[phase]
	if !ok {
		value = 0
	}
	releasePhase.With(
		LabelNamespace, namespace,
		LabelReleaseName, releaseName,
	).Set(value)
}
