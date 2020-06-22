package status

import (
	v1 "github.com/fluxcd/helm-operator/pkg/apis/helm.fluxcd.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestHasTestsFailed(t *testing.T) {
	type args struct {
		hr *v1.HelmRelease
	}
	var tests = []struct {
		name string
		args args
		want bool
	}{
		{
			name: "should return false when helm release has not synced",
			args: args{hr: &v1.HelmRelease{
				ObjectMeta: defaultSyncedObjectMeta(),
				Status:     defaultSyncedReleaseStatus([]v1.HelmReleaseCondition{}),
			}},
		}, {
			name: "should return false when helm release does not have condition 'HelmReleaseTested'",
			args: args{hr: &v1.HelmRelease{
				ObjectMeta: defaultSyncedObjectMeta(),
				Status:     defaultSyncedReleaseStatus([]v1.HelmReleaseCondition{}),
			}},
		}, {
			name: "should return true when helm release has condition 'HelmReleaseTested' that has status 'ConditionFalse'",
			args: args{hr: &v1.HelmRelease{
				ObjectMeta: defaultSyncedObjectMeta(),
				Status: defaultSyncedReleaseStatus(
					[]v1.HelmReleaseCondition{{
						Type:   v1.HelmReleaseTested,
						Status: v1.ConditionFalse,
					}}),
			}},
			want: true,
		}, {
			name: "should return false when helm release has condition 'HelmReleaseTested' and has status 'ConditionTrue'",
			args: args{hr: &v1.HelmRelease{
				ObjectMeta: defaultSyncedObjectMeta(),
				Status: defaultSyncedReleaseStatus(
					[]v1.HelmReleaseCondition{{
						Type:   v1.HelmReleaseTested,
						Status: v1.ConditionTrue,
					}}),
			}},
		}, {
			name: "should return false when helm release has condition 'HelmReleaseTested' and has status 'ConditionUnknown'",
			args: args{hr: &v1.HelmRelease{
				ObjectMeta: defaultSyncedObjectMeta(),
				Status: defaultSyncedReleaseStatus(
					[]v1.HelmReleaseCondition{{
						Type:   v1.HelmReleaseTested,
						Status: v1.ConditionUnknown,
					}}),
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasTestsFailed(tt.args.hr); got != tt.want {
				t.Errorf("HasTestsFailed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func defaultSyncedReleaseStatus(conditions []v1.HelmReleaseCondition) v1.HelmReleaseStatus {
	return v1.HelmReleaseStatus{
		ObservedGeneration: int64(4),
		Conditions:         conditions,
	}
}

func defaultSyncedObjectMeta() metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Generation: int64(2),
	}
}
