package v2

import (
	"fmt"

	helmv2 "k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/proto/hapi/release"

	"github.com/fluxcd/helm-operator/pkg/helm"
)

func (h *HelmV2) Test(releaseName string, opts helm.TestOptions) error {
	c, errc := h.client.RunReleaseTest(
		releaseName,
		helmv2.ReleaseTestCleanup(opts.Cleanup),
		helmv2.ReleaseTestTimeout(int64(opts.Timeout.Seconds())),
	)

	testErr := &testErr{}

	for {
		select {
		case err := <-errc:
			if err != nil {
				return statusMessageErr(err)
			}
			if testErr.failed > 0 {
				return testErr.Error()
			}
			return nil
		case res, ok := <-c:
			if !ok {
				break
			}

			if res.Status == release.TestRun_FAILURE {
				testErr.failed++
			}

			h.logger.Log("info", res.Msg)
		}
	}
}

type testErr struct {
	failed int
}

func (err *testErr) Error() error {
	return fmt.Errorf("%v test(s) failed", err.failed)
}
