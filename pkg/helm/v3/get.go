package v3

import (
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/storage/driver"

	"github.com/fluxcd/helm-operator/pkg/helm"
)

type (
	getOptions    helm.GetOptions
	statusOptions helm.GetOptions
)

func (h *HelmV3) Get(releaseName string, opts helm.GetOptions) (*helm.Release, error) {
	cfg, err := newActionConfig(h.kubeConfig, h.infoLogFunc(opts.Namespace, releaseName), opts.Namespace, "")
	if err != nil {
		return nil, err
	}

	get := action.NewGet(cfg)
	getOptions(opts).configure(get)

	res, err := get.Run(releaseName)
	switch err {
	case nil:
		return releaseToGenericRelease(res), nil
	case driver.ErrReleaseNotFound:
		return nil, nil
	default:
		return nil, err
	}
}

func (opts getOptions) configure(action *action.Get) {
	action.Version = opts.Version
}

func (h *HelmV3) Status(releaseName string, opts helm.StatusOptions) (helm.Status, error) {
	cfg, err := newActionConfig(h.kubeConfig, h.infoLogFunc(opts.Namespace, releaseName), opts.Namespace, "")
	if err != nil {
		return "", err
	}

	status:= action.NewStatus(cfg)
	statusOptions(opts).configure(status)

	res, err := status.Run(releaseName)
	switch err {
	case nil:
		return lookUpGenericStatus(res.Info.Status), nil
	default:
		return "", err
	}
}

func (opts statusOptions) configure(action *action.Status) {
	action.Version = opts.Version
}
