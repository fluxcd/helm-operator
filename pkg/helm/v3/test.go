package v3

import (
	"helm.sh/helm/v3/pkg/action"

	"github.com/fluxcd/helm-operator/pkg/helm"
)

func (h *HelmV3) Test(releaseName string, opts helm.TestOptions) error {
	cfg, err := newActionConfig(h.kubeConfig, h.infoLogFunc(opts.Namespace, releaseName), opts.Namespace, "")
	if err != nil {
		return err
	}

	test := action.NewReleaseTesting(cfg)
	testOptions(opts).configure(test)

	if _, err := test.Run(releaseName); err != nil {
		return err
	}

	return nil
}

type testOptions helm.TestOptions

func (opts testOptions) configure(action *action.ReleaseTesting) {
	action.Timeout = opts.Timeout
}
