package helm

import (
	"time"
)

type RollbackOptions struct {
	Namespace    string        `json:"namespace,omitempty"`
	Version      int           `json:"version,omitempty"`
	Timeout      time.Duration `json:"timeout,omitempty"`
	Wait         bool          `json:"wait,omitempty"`
	DisableHooks bool          `json:"disableHooks,omitempty"`
	DryRun       bool          `json:"dryRun,omitempty"`
	Recreate     bool          `json:"recreate,omitempty"`
	Force        bool          `json:"force,omitempty"`
}
