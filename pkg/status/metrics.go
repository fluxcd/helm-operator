package status

import (
	v1 "github.com/fluxcd/helm-operator/pkg/apis/helm.fluxcd.io/v1"
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
	releaseCondition = stdprometheus.NewGaugeVec(stdprometheus.GaugeOpts{
		Namespace: "flux",
		Subsystem: "helm_operator",
		Name:      "release_condition_info",
		Help:      "Current HelmRelease condition status. Values are -1 (false), 0 (unknown or absent), 1 (true)",
	}, []string{LabelTargetNamespace, LabelReleaseName, LabelCondition})
)

func init() {
	stdprometheus.MustRegister(releaseCondition)
}

func ObserveReleaseConditions(old *v1.HelmRelease, new *v1.HelmRelease) {
	conditions := make(map[v1.HelmReleaseConditionType]*v1.ConditionStatus)

	for _, condition := range old.Status.Conditions {
		conditions[condition.Type] = nil
	}

	if new != nil {
		for _, condition := range new.Status.Conditions {
			conditions[condition.Type] = &condition.Status
		}
	}

	for conditionType, conditionStatus := range conditions {
		if conditionStatus == nil {
			releaseCondition.Delete(labelsForRelease(old, conditionType))
		} else {
			releaseCondition.With(labelsForRelease(new, conditionType)).Set(conditionStatusToGaugeValue[*conditionStatus])
		}
	}
}

func labelsForRelease(hr *v1.HelmRelease, conditionType v1.HelmReleaseConditionType) stdprometheus.Labels {
	return stdprometheus.Labels{
		LabelTargetNamespace: hr.GetTargetNamespace(),
		LabelReleaseName:     hr.GetReleaseName(),
		LabelCondition:       string(conditionType),
	}
}
