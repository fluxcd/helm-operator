package status

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"k8s.io/utils/clock"

	"github.com/fluxcd/helm-operator/pkg/apis/helm.fluxcd.io/v1"
	v1client "github.com/fluxcd/helm-operator/pkg/client/clientset/versioned/typed/helm.fluxcd.io/v1"
)

// Clock is defined as a var so it can be stubbed during tests.
var Clock clock.Clock = clock.RealClock{}

func GetCondition(status v1.HelmReleaseStatus, conditionType v1.HelmReleaseConditionType) *v1.HelmReleaseCondition {
	for i := range status.Conditions {
		c := status.Conditions[i]
		if c.Type == conditionType {
			return &c
		}
	}
	return nil
}

func SetCondition(client v1client.HelmReleaseInterface, hr *v1.HelmRelease, condition v1.HelmReleaseCondition, set func(*v1.HelmRelease)) error {
	firstTry := true
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
		if !firstTry {
			var getErr error
			hr, getErr = client.Get(hr.Name, metav1.GetOptions{})
			if getErr != nil {
				return getErr
			}
		}

		currCondition := GetCondition(hr.Status, condition.Type)
		if currCondition != nil && currCondition.Status == condition.Status {
			condition.LastTransitionTime = currCondition.LastTransitionTime
		}

		cHr := hr.DeepCopy()
		cHr.Status.Conditions = append(filterOutCondition(cHr.Status.Conditions, condition.Type), condition)
		switch {
		case condition.Type == v1.HelmReleaseReleased && condition.Status == v1.ConditionTrue:
			cHr.Status.Conditions = filterOutCondition(cHr.Status.Conditions, v1.HelmReleaseRolledBack)
			cHr.Status.RollbackCount = 0
		case condition.Type == v1.HelmReleaseRolledBack && condition.Status == v1.ConditionTrue:
			cHr.Status.RollbackCount = cHr.Status.RollbackCount + 1
		}
		if set != nil {
			set(cHr)
		}

		_, err = client.UpdateStatus(cHr)
		firstTry = false
		return
	})
	return err
}

func SetStatusPhase(client v1client.HelmReleaseInterface, hr *v1.HelmRelease, phase v1.HelmReleasePhase) error {
	condition, ok := ConditionForPhase(hr, phase)
	if !ok {
		return nil
	}
	return SetCondition(client, hr, condition, func(cHr *v1.HelmRelease) {
		cHr.Status.Phase = phase
	})
}

func SetStatusPhaseWithRevision(client v1client.HelmReleaseInterface, hr *v1.HelmRelease, phase v1.HelmReleasePhase, revision string) error {
	condition, ok := ConditionForPhase(hr, phase)
	if !ok {
		return nil
	}
	return SetCondition(client, hr, condition, func(cHr *v1.HelmRelease) {
		switch {
		case phase == v1.HelmReleasePhaseInstalling || phase == v1.HelmReleasePhaseUpgrading:
			cHr.Status.LastAttemptedRevision = revision
		case phase == v1.HelmReleasePhaseSucceeded:
			cHr.Status.Revision = revision
		}
		cHr.Status.Phase = phase
	})
}

// ConditionForPhrase returns a condition for the given phase.
func ConditionForPhase(hr *v1.HelmRelease, phase v1.HelmReleasePhase) (v1.HelmReleaseCondition, bool) {
	nowTime := metav1.NewTime(Clock.Now())
	condition := &v1.HelmReleaseCondition{
		Reason:             string(phase),
		LastUpdateTime:     &nowTime,
		LastTransitionTime: &nowTime,
	}
	switch phase {
	case v1.HelmReleasePhaseInstalling:
		condition.Type = v1.HelmReleaseReleased
		condition.Status = v1.ConditionUnknown
		condition.Message = fmt.Sprintf(`Running installation for Helm release '%s' in '%s'.`, hr.GetReleaseName(), hr.GetTargetNamespace())
	case v1.HelmReleasePhaseUpgrading:
		condition.Type = v1.HelmReleaseReleased
		condition.Status = v1.ConditionUnknown
		condition.Message = fmt.Sprintf(`Running upgrade for Helm release '%s' in '%s'.`, hr.GetReleaseName(), hr.GetTargetNamespace())
	case v1.HelmReleasePhaseSucceeded:
		condition.Type = v1.HelmReleaseReleased
		condition.Status = v1.ConditionTrue
		condition.Message = fmt.Sprintf(`Release was successful for Helm release '%s' in '%s'.`, hr.GetReleaseName(), hr.GetTargetNamespace())
	case v1.HelmReleasePhaseFailed:
		condition.Type = v1.HelmReleaseReleased
		condition.Status = v1.ConditionFalse
		condition.Message = fmt.Sprintf(`Release failed for Helm release '%s' in '%s'.`, hr.GetReleaseName(), hr.GetTargetNamespace())
	case v1.HelmReleasePhaseRollingBack:
		condition.Type = v1.HelmReleaseRolledBack
		condition.Status = v1.ConditionUnknown
		condition.Message = fmt.Sprintf(`Rolling back Helm release '%s' in '%s'.`, hr.GetReleaseName(), hr.GetTargetNamespace())
	case v1.HelmReleasePhaseRolledBack:
		condition.Type = v1.HelmReleaseRolledBack
		condition.Status = v1.ConditionTrue
		condition.Message = fmt.Sprintf(`Rolled back Helm release '%s' in '%s'.`, hr.GetReleaseName(), hr.GetTargetNamespace())
	case v1.HelmReleasePhaseRollbackFailed:
		condition.Type = v1.HelmReleaseRolledBack
		condition.Status = v1.ConditionFalse
		condition.Message = fmt.Sprintf(`Rollback failed for Helm release '%s' in '%s'.`, hr.GetReleaseName(), hr.GetTargetNamespace())
	case v1.HelmReleasePhaseChartFetched:
		condition.Type = v1.HelmReleaseChartFetched
		condition.Status = v1.ConditionTrue
		condition.Message = fmt.Sprintf(`Chart fetch was successful for Helm release '%s' in '%s'.`, hr.GetReleaseName(), hr.GetTargetNamespace())
	case v1.HelmReleasePhaseChartFetchFailed:
		condition.Type = v1.HelmReleaseChartFetched
		condition.Status = v1.ConditionFalse
		condition.Message = fmt.Sprintf(`Chart fetch failed for Helm release '%s' in '%s'.`, hr.GetReleaseName(), hr.GetTargetNamespace())
	default:
		return v1.HelmReleaseCondition{}, false
	}
	return *condition, true
}

// filterOutCondition returns a new slice of condition without the
// condition of the given type.
func filterOutCondition(conditions []v1.HelmReleaseCondition,
	conditionType v1.HelmReleaseConditionType) []v1.HelmReleaseCondition {

	var newConditions []v1.HelmReleaseCondition
	for _, c := range conditions {
		if c.Type == conditionType {
			continue
		}
		newConditions = append(newConditions, c)
	}
	return newConditions
}
