package release

import (
	"strings"
)

type errCollection []error

func (err errCollection) Error() string {
	var errs []string
	for i := len(err)-1; i >= 0; i-- {
		errs = append(errs, err[i].Error())
	}
	return strings.Join(errs, ", previous error:")
}

func (err errCollection) Empty() bool {
	return len(err) == 0
}
