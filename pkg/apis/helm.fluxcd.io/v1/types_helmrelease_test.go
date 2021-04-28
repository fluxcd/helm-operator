package v1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
