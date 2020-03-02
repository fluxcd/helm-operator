package v1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHelmValues(t *testing.T) {
	testCases := []struct {
		original         *HelmValues
		transformer      func(v *HelmValues) *HelmValues
		expectedCopy     *HelmValues
		expectedOriginal *HelmValues
	}{
		{
			original: &HelmValues{Data: map[string]interface{}{}},
			transformer: func(v *HelmValues) *HelmValues {
				v.Data["foo"] = "bar"
				return v
			},
			expectedCopy:     &HelmValues{Data: map[string]interface{}{"foo": "bar"}},
			expectedOriginal: &HelmValues{Data: map[string]interface{}{}},
		},
		{
			original: &HelmValues{Data: map[string]interface{}{"foo": map[string]interface{}{"bar": "baz"}}},
			transformer: func(v *HelmValues) *HelmValues {
				v.Data["foo"] = map[string]interface{}{"bar": "oof"}
				return v
			},
			expectedCopy:     &HelmValues{Data: map[string]interface{}{"foo": map[string]interface{}{"bar": "oof"}}},
			expectedOriginal: &HelmValues{Data: map[string]interface{}{"foo": map[string]interface{}{"bar": "baz"}}},
		},
	}

	for i, tc := range testCases {
		var out HelmValues
		tc.original.DeepCopyInto(&out)
		assert.Exactly(t, tc.expectedCopy, tc.transformer(&out), "copy was not mutated. test case: %d", i)
		assert.Exactly(t, tc.expectedOriginal, tc.original, "original was mutated. test case: %d", i)
	}
}

func TestRefOrDefault(t *testing.T) {
	testCases := []struct {
		chartSource      GitChartSource
		potentialDefault string
		expected         string
	}{
		{
			chartSource: GitChartSource{
				Ref: "master",
			},
			potentialDefault: "dev",
			expected:         "master",
		},
		{
			chartSource:      GitChartSource{},
			potentialDefault: "dev",
			expected:         "dev",
		},
	}

	for _, tc := range testCases {
		got := tc.chartSource.RefOrDefault(tc.potentialDefault)
		assert.Equal(t, tc.expected, got)
	}
}
