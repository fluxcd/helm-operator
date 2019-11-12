package v2

import (
	"github.com/pkg/errors"
	"google.golang.org/grpc/status"

	helmv2 "k8s.io/helm/pkg/helm"

	"github.com/fluxcd/helm-operator/pkg/helm"
)

func (h *HelmV2) Uninstall(releaseName string, opts helm.UninstallOptions) error {
	if _, err := h.client.DeleteRelease(
		releaseName,
		helmv2.DeleteDisableHooks(opts.DisableHooks),
		helmv2.DeleteDryRun(opts.DryRun),
		helmv2.DeletePurge(!opts.KeepHistory),
		helmv2.DeleteTimeout(int64(opts.Timeout.Seconds())),
	); err != nil {
		if s, ok := status.FromError(err); ok {
			return errors.New(s.Message())
		}
		return err
	}
	return nil
}
