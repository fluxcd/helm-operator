package helm

import (
	"reflect"

	"github.com/google/go-cmp/cmp"
)

func Diff(j *Release, k *Release) string {
	opt := cmp.FilterValues(func(x, y interface{}) bool {
		isNumeric := func(v interface{}) bool {
			return v != nil && reflect.TypeOf(v).ConvertibleTo(reflect.TypeOf(float64(0)))
		}
		return isNumeric(x) && isNumeric(y)
	}, cmp.Transformer("T", func(v interface{}) float64 {
		return reflect.ValueOf(v).Convert(reflect.TypeOf(float64(0))).Float()
	}))

	return cmp.Diff(j.Values, k.Values, opt) + cmp.Diff(j.Chart, k.Chart, opt)
}
