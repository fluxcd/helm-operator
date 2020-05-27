package status

import (
	v1 "github.com/fluxcd/helm-operator/pkg/apis/helm.fluxcd.io/v1"
	"github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

const (
	LabelTargetNamespace = "target_namespace"
	LabelReleaseName     = "release_name"
	LabelCondition       = "condition"
)

var (
	conditionStatusToGaugeValue = map[v1.ConditionStatus]float64{
		v1.ConditionFalse:   -1,
		v1.ConditionUnknown: 0,
		v1.ConditionTrue:    1,
	}
	releaseCondition = prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
		Namespace: "flux",
		Subsystem: "helm_operator",
		Name:      "release_condition_info",
		Help:      "Current HelmRelease condition status. Values are -1 (false), 0 (unknown or absent), 1 (true)",
	}, []string{LabelTargetNamespace, LabelReleaseName, LabelCondition})
)

func ObserveReleaseConditions(old v1.HelmRelease, new v1.HelmRelease) {
	conditions := make(map[v1.HelmReleaseConditionType]v1.ConditionStatus)
	for _, condition := range old.Status.Conditions {
		// Initialize conditions from old status to unknown, so that if
		// they are removed in new status, they do not contain stale data.
		conditions[condition.Type] = v1.ConditionUnknown
	}
	for _, condition := range new.Status.Conditions {
		conditions[condition.Type] = condition.Status
	}
	for conditionType, conditionStatus := range conditions {
		releaseCondition.With(
			LabelTargetNamespace, new.GetTargetNamespace(),
			LabelReleaseName, new.GetReleaseName(),
			LabelCondition, string(conditionType),
		).Set(conditionStatusToGaugeValue[conditionStatus])
	}
}
