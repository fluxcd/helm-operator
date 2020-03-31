package helm

import (
	"github.com/google/go-cmp/cmp"
)

func Diff(j *Release, k *Release) string {
	return cmp.Diff(j.Values, k.Values) + cmp.Diff(j.Chart, k.Chart)
}
