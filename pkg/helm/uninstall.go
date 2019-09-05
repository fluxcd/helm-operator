package helm

import (
	"time"
)

type UninstallOptions struct {
	Namespace    string        `json:"namespace,omitempty"`
	DisableHooks bool          `json:"disableHooks,omitempty"`
	DryRun       bool          `json:"dryRun,omitempty"`
	KeepHistory  bool          `json:"keepHistory,omitempty"`
	Timeout      time.Duration `json:"timeout,omitempty"`
}
