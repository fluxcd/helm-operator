/*

This package is for maintaining the link between `HelmRelease`
resources and the Helm releases to which they
correspond. Specifically,

 1. updating the `HelmRelease` status based on the progress of
   syncing, and the state of the associated Helm release; and,

 2. attributing each resource in a Helm release (under our control) to
 the associated `HelmRelease`.

*/
package status

import (
	"time"

	"github.com/go-kit/kit/log"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	kube "k8s.io/client-go/kubernetes"

	helmfluxv1 "github.com/fluxcd/helm-operator/pkg/apis/helm.fluxcd.io/v1"
	ifclientset "github.com/fluxcd/helm-operator/pkg/client/clientset/versioned"
	v1client "github.com/fluxcd/helm-operator/pkg/client/clientset/versioned/typed/helm.fluxcd.io/v1"
	iflister "github.com/fluxcd/helm-operator/pkg/client/listers/helm.fluxcd.io/v1"
	"github.com/fluxcd/helm-operator/pkg/helm"
)

const period = 10 * time.Second

type Updater struct {
	hrClient    ifclientset.Interface
	hrLister    iflister.HelmReleaseLister
	kube        kube.Interface
	helmClients *helm.Clients
	namespace   string
}

func New(hrClient ifclientset.Interface, hrLister iflister.HelmReleaseLister, helmClients *helm.Clients) *Updater {
	return &Updater{
		hrClient:    hrClient,
		hrLister:    hrLister,
		helmClients: helmClients,
	}
}

func (u *Updater) Loop(stop <-chan struct{}, logger log.Logger) {
	ticker := time.NewTicker(period)
	var logErr error

bail:
	for {
		select {
		case <-stop:
			break bail
		case <-ticker.C:
		}
		list, err := u.hrLister.List(labels.Everything())
		if err != nil {
			logErr = err
			break bail
		}
		for _, hr := range list {
			nsHrClient := u.hrClient.HelmV1().HelmReleases(hr.Namespace)
			releaseName := hr.GetReleaseName()
			c, ok := u.helmClients.Load(hr.GetHelmVersion())
			// If we are unable to get the client, we do not care why
			if !ok {
				continue
			}
			rel, err := c.Status(hr.GetReleaseName(), helm.StatusOptions{})
			// If we are unable to get the status, we do not care why
			if err != nil {
				continue
			}
			if err := SetReleaseStatus(nsHrClient, *hr, releaseName, rel.Info.Status.String()); err != nil {
				logger.Log("namespace", hr.Namespace, "resource", hr.Name, "err", err)
				continue
			}
		}
	}

	ticker.Stop()
	logger.Log("loop", "stopping", "err", logErr)
}

// SetReleaseStatus updates the status of the HelmRelease to the given
// release name and/or release status.
func SetReleaseStatus(client v1client.HelmReleaseInterface, hr helmfluxv1.HelmRelease,
	releaseName, releaseStatus string) error {

	cHr, err := client.Get(hr.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if cHr.Status.ReleaseName == releaseName && cHr.Status.ReleaseStatus == releaseStatus {
		return nil
	}

	cHr.Status.ReleaseName = releaseName
	cHr.Status.ReleaseStatus = releaseStatus

	_, err = client.UpdateStatus(cHr)
	return err
}

// SetReleaseRevision updates the status of the HelmRelease to the
// given revision.
func SetReleaseRevision(client v1client.HelmReleaseInterface, hr helmfluxv1.HelmRelease, revision string) error {
	cHr, err := client.Get(hr.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if cHr.Status.Revision == revision {
		return nil
	}

	cHr.Status.Revision = revision

	_, err = client.UpdateStatus(cHr)
	return err
}

// SetValuesChecksum updates the values checksum of the HelmRelease to
// the given checksum.
func SetValuesChecksum(client v1client.HelmReleaseInterface, hr helmfluxv1.HelmRelease, valuesChecksum string) error {
	cHr, err := client.Get(hr.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if valuesChecksum == "" || cHr.Status.ValuesChecksum == valuesChecksum {
		return nil
	}

	cHr.Status.ValuesChecksum = valuesChecksum

	_, err = client.UpdateStatus(cHr)
	return err
}

// SetObservedGeneration updates the observed generation status of the
// HelmRelease to the given generation.
func SetObservedGeneration(client v1client.HelmReleaseInterface, hr helmfluxv1.HelmRelease, generation int64) error {
	cHr, err := client.Get(hr.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if cHr.Status.ObservedGeneration >= generation {
		return nil
	}

	cHr.Status.ObservedGeneration = generation

	_, err = client.UpdateStatus(cHr)
	return err
}

// HasSynced returns if the HelmRelease has been processed by the
// controller.
func HasSynced(hr helmfluxv1.HelmRelease) bool {
	return hr.Status.ObservedGeneration >= hr.Generation
}

// HasRolledBack returns if the current generation of the HelmRelease
// has been rolled back.
func HasRolledBack(hr helmfluxv1.HelmRelease) bool {
	if !HasSynced(hr) {
		return false
	}

	rolledBack := GetCondition(hr.Status, helmfluxv1.HelmReleaseRolledBack)
	if rolledBack == nil {
		return false
	}

	chartFetched := GetCondition(hr.Status, helmfluxv1.HelmReleaseChartFetched)
	if chartFetched != nil {
		// NB: as two successful state updates can happen right after
		// each other, on which we both want to act, we _must_ compare
		// the update timestamps as the transition timestamp will only
		// change on a status shift.
		if chartFetched.Status == v1.ConditionTrue && rolledBack.LastUpdateTime.Before(&chartFetched.LastUpdateTime) {
			return false
		}
	}

	return rolledBack.Status == v1.ConditionTrue
}
