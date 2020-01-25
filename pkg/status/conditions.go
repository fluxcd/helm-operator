package status

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"

	helmfluxv1 "github.com/fluxcd/helm-operator/pkg/apis/helm.fluxcd.io/v1"
	v1client "github.com/fluxcd/helm-operator/pkg/client/clientset/versioned/typed/helm.fluxcd.io/v1"
)

// NewCondition creates a new HelmReleaseCondition.
func NewCondition(conditionType helmfluxv1.HelmReleaseConditionType, status v1.ConditionStatus,
	reason, message string) helmfluxv1.HelmReleaseCondition {

	return helmfluxv1.HelmReleaseCondition{
		Type:               conditionType,
		Status:             status,
		LastUpdateTime:     metav1.Now(),
		LastTransitionTime: metav1.Now(),
		Reason:             reason,
		Message:            message,
	}
}

// SetCondition updates the HelmRelease to include the given condition.
func SetCondition(client v1client.HelmReleaseInterface, hr *helmfluxv1.HelmRelease,
	condition helmfluxv1.HelmReleaseCondition) error {

	firstTry := true
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
		if !firstTry {
			var getErr error
			hr, getErr = client.Get(hr.Name, metav1.GetOptions{})
			if getErr != nil {
				return getErr
			}
		}

		cHr := hr.DeepCopy()
		currCondition := GetCondition(cHr.Status, condition.Type)
		if currCondition != nil && currCondition.Status == condition.Status {
			condition.LastTransitionTime = currCondition.LastTransitionTime
		}
		newConditions := filterOutCondition(cHr.Status.Conditions, condition.Type)
		cHr.Status.Conditions = append(newConditions, condition)

		_, err = client.UpdateStatus(cHr)
		firstTry = false
		return
	})
	return err
}

// GetCondition returns the condition with the given type.
func GetCondition(status helmfluxv1.HelmReleaseStatus,
	conditionType helmfluxv1.HelmReleaseConditionType) *helmfluxv1.HelmReleaseCondition {

	for i := range status.Conditions {
		c := status.Conditions[i]
		if c.Type == conditionType {
			return &c
		}
	}
	return nil
}

// filterOutCondition returns a new slice of conditions without the
// conditions of the given type.
func filterOutCondition(conditions []helmfluxv1.HelmReleaseCondition,
	conditionType helmfluxv1.HelmReleaseConditionType) []helmfluxv1.HelmReleaseCondition {

	var newConditions []helmfluxv1.HelmReleaseCondition
	for _, c := range conditions {
		if c.Type == conditionType {
			continue
		}
		newConditions = append(newConditions, c)
	}
	return newConditions
}
