package helm

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

type UpgradeOptions struct {
	Namespace    string        `json:"namespace,omitempty"`
	Timeout      time.Duration `json:"timeout,omitempty"`
	Wait         bool          `json:"wait,omitempty"`
	DisableHooks bool          `json:"disableHooks,omitempty"`
	DryRun       bool          `json:"dryRun,omitempty"`
	Force        bool          `json:"force,omitempty"`
	ResetValues  bool          `json:"resetValues,omitempty"`
	ReuseValues  bool          `json:"reuseValues,omitempty"`
	Recreate     bool          `json:"recreate,omitempty"`
	MaxHistory   int           `json:"maxHistory,omitempty"`
	Atomic       bool          `json:"atomic,omitempty"`
}

func (o UpgradeOptions) Validate(unsupported []string) error {
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
