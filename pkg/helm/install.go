package helm

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

type InstallOptions struct {
	Namespace        string        `json:"namespace,omitempty"`
	ClientOnly       bool          `json:"clientOnly,omitempty"`
	DryRun           bool          `json:"dryRun,omitempty"`
	DisableHooks     bool          `json:"disableHooks,omitempty"`
	DisableCRDHooks  bool          `json:"disableCRDHooks,omitempty"`
	Replace          bool          `json:"replace,omitempty"`
	Wait             bool          `json:"wait,omitempty"`
	DependencyUpdate bool          `json:"dependencyUpdate,omitempty"`
	Timeout          time.Duration `json:"timeout,omitempty"`
	Atomic           bool          `json:"atomic,omitempty"`
}

func (o InstallOptions) Validate(unsupported []string) error {
	var optionMap map[string]interface{}
	b, _ := json.Marshal(o)
	json.Unmarshal(b, &optionMap)

	var setUnsupportedOpts []string
	for _, k := range unsupported {
		if _, ok := optionMap[k]; ok {
			setUnsupportedOpts = append(setUnsupportedOpts, k)
		}
	}
	if len(setUnsupportedOpts) > 0 {
		return errors.New(fmt.Sprintf(
			"configuring any of {%s} has no effect for this Client version",
			strings.Join(unsupported, ", "),
		))
	}
	return nil
}
