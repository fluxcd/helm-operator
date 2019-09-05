/*

This package has the algorithm for making sure the Helm releases in
the cluster match what are defined in the HelmRelease resources.

There are several ways they can be mismatched. Here's how they are
reconciled:

 1a. There is a HelmRelease resource, but no corresponding
   release. This can happen when the helm operator is first run, for
   example.

 1b. The release corresponding to a HelmRelease has been updated by
   some other means, perhaps while the operator wasn't running. This
   is also checked, by doing a dry-run release and comparing the result
   to the release.

 2. The chart has changed in git, meaning the release is out of
   date. The ChartChangeSync responds to new git commits by looking up
   each chart that makes use of the mirror that has new commits,
   replacing the clone for that chart, and scheduling a new release.

1a.) and 1b.) run on the same schedule, and 2.) is run when a git
mirror reports it has fetched from upstream _and_ (upon checking) the
head of the branch has changed.

Since both 1*.) and 2.) look at the charts in the git repo, but run on
different schedules (non-deterministically), there's a chance that
they can fight each other. For example, the git mirror may fetch new
commits which are used in 1), then treated as changes subsequently by
2). To keep consistency between the two, the current revision of a
repo is used by 1), and advanced only by 2).

*/
package chartsync

import (
	"context"
	"fmt"
	"github.com/fluxcd/helm-operator/pkg/helm"
	"path/filepath"
	"sync"
	"time"

	helmfluxv1 "github.com/fluxcd/helm-operator/pkg/apis/helm.fluxcd.io/v1"
	ifclientset "github.com/fluxcd/helm-operator/pkg/client/clientset/versioned"
	iflister "github.com/fluxcd/helm-operator/pkg/client/listers/helm.fluxcd.io/v1"
	"github.com/fluxcd/helm-operator/pkg/release"
	"github.com/fluxcd/helm-operator/pkg/status"
	"github.com/go-kit/kit/log"
	"github.com/google/go-cmp/cmp"
	"github.com/weaveworks/flux/git"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

const (
	// condition change reasons
	ReasonGitNotReady      = "GitRepoNotCloned"
	ReasonDownloadFailed   = "RepoFetchFailed"
	ReasonDownloaded       = "RepoChartInCache"
	ReasonInstallFailed    = "HelmInstallFailed"
	ReasonDependencyFailed = "UpdateDependencyFailed"
	ReasonUpgradeFailed    = "HelmUpgradeFailed"
	ReasonRollbackFailed   = "HelmRollbackFailed"
	ReasonCloned           = "GitRepoCloned"
	ReasonSuccess          = "HelmSuccess"
)

type Clients struct {
	KubeClient  kubernetes.Clientset
	IfClient    ifclientset.Clientset
	HrLister    iflister.HelmReleaseLister
	HelmClients *helm.Clients
}

type Config struct {
	ChartCache      string
	LogDiffs        bool
	UpdateDeps      bool
	GitTimeout      time.Duration
	GitPollInterval time.Duration
}

func (c Config) WithDefaults() Config {
	if c.ChartCache == "" {
		c.ChartCache = "/tmp"
	}
	return c
}

// clone puts a local git clone together with its state (head
// revision), so we can keep track of when it needs to be updated.
type clone struct {
	export *git.Export
	remote string
	ref    string
	head   string
}

// ReleaseQueue is an add-only workqueue.RateLimitingInterface
type ReleaseQueue interface {
	AddRateLimited(item interface{})
}

type ChartChangeSync struct {
	logger       log.Logger
	kubeClient   kubernetes.Clientset
	ifClient     ifclientset.Clientset
	hrLister     iflister.HelmReleaseLister
	helmClients  *helm.Clients
	releaseQueue ReleaseQueue
	config       Config

	mirrors *git.Mirrors

	clonesMu sync.Mutex
	clones   map[string]clone

	namespace string
}

func New(logger log.Logger, clients Clients, releaseQueue ReleaseQueue, config Config, namespace string) *ChartChangeSync {
	return &ChartChangeSync{
		logger:       logger,
		kubeClient:   clients.KubeClient,
		ifClient:     clients.IfClient,
		hrLister:     clients.HrLister,
		helmClients:  clients.HelmClients,
		releaseQueue: releaseQueue,
		config:       config.WithDefaults(),
		mirrors:      git.NewMirrors(),
		clones:       make(map[string]clone),
		namespace:    namespace,
	}
}

// Run creates a syncing loop that will reconcile differences between
// Helm releases in the cluster, what HelmRelease declare, and
// changes in the git repos mentioned by any HelmRelease.
func (chs *ChartChangeSync) Run(stopCh <-chan struct{}, errc chan error, wg *sync.WaitGroup) {
	chs.logger.Log("info", "starting git chart sync loop")

	wg.Add(1)
	go func() {
		defer runtime.HandleCrash()
		defer func() {
			chs.mirrors.StopAllAndWait()
			wg.Done()
		}()

		for {
			select {
			case mirrorsChanged := <-chs.mirrors.Changes():
				for mirror := range mirrorsChanged {
					resources, err := chs.getCustomResourcesForMirror(mirror)
					if err != nil {
						chs.logger.Log("warning", "failed to get custom resources", "err", err)
						continue
					}

					// Retrieve the mirror we got a change signal for
					repo, ok := chs.mirrors.Get(mirror)
					if !ok {
						// Then why .. did you say .. it had changed? It may have been removed. Add it back and let it signal again.
						chs.logger.Log("warning", "mirrored git repo disappeared after signalling change", "repo", mirror)
						for _, hr := range resources {
							chs.setCondition(hr, helmfluxv1.HelmReleaseChartFetched, v1.ConditionUnknown, ReasonGitNotReady, "git mirror missing; starting mirroring again")
							chs.maybeMirror(hr)
						}
						continue
					}

					// Ensure the repo is ready
					status, err := repo.Status()
					if status != git.RepoReady {
						chs.logger.Log("info", "repo not ready yet, while attempting chart sync", "repo", mirror, "status", string(status))
						for _, hr := range resources {
							// TODO(michael) log if there's a problem with the following?
							chs.setCondition(hr, helmfluxv1.HelmReleaseChartFetched, v1.ConditionUnknown, ReasonGitNotReady, err.Error())
						}
						continue
					}

					// Determine if we need to update the clone and
					// schedule an upgrade for every HelmRelease that
					// makes use of the mirror
					for _, hr := range resources {
						ref := hr.Spec.ChartSource.GitChartSource.RefOrDefault()
						path := hr.Spec.ChartSource.GitChartSource.Path
						releaseName := hr.GetReleaseName()

						ctx, cancel := context.WithTimeout(context.Background(), chs.config.GitTimeout)
						refHead, err := repo.Revision(ctx, ref)
						cancel()
						if err != nil {
							chs.setCondition(hr, helmfluxv1.HelmReleaseChartFetched, v1.ConditionFalse, ReasonGitNotReady, "problem cloning from local git mirror: "+err.Error())
							chs.logger.Log("warning", "could not get revision for ref while checking for changes", "resource", hr.ResourceID().String(), "repo", mirror, "ref", ref, "err", err)
							continue
						}

						// The git repo of this appears to have had commits since we last saw it,
						// check explicitly whether we should update its clone.
						chs.clonesMu.Lock()
						cloneForChart, ok := chs.clones[releaseName]
						chs.clonesMu.Unlock()

						if ok { // found clone
							ctx, cancel := context.WithTimeout(context.Background(), chs.config.GitTimeout)
							commits, err := repo.CommitsBetween(ctx, cloneForChart.head, refHead, path)
							cancel()
							if err != nil {
								chs.setCondition(hr, helmfluxv1.HelmReleaseChartFetched, v1.ConditionFalse, ReasonGitNotReady, "problem cloning from local git mirror: "+err.Error())
								chs.logger.Log("warning", "could not get revision for ref while checking for changes", "resource", hr.ResourceID().String(), "repo", mirror, "ref", ref, "err", err)
								continue
							}
							ok = len(commits) == 0
						}

						if !ok { // didn't find clone, or it needs updating
							ctx, cancel := context.WithTimeout(context.Background(), chs.config.GitTimeout)
							newClone, err := repo.Export(ctx, refHead)
							cancel()
							if err != nil {
								chs.setCondition(hr, helmfluxv1.HelmReleaseChartFetched, v1.ConditionFalse, ReasonGitNotReady, "problem cloning from local git mirror: "+err.Error())
								chs.logger.Log("warning", "could not clone from mirror while checking for changes", "resource", hr.ResourceID().String(), "repo", mirror, "ref", ref, "err", err)
								continue
							}
							newCloneForChart := clone{remote: mirror, ref: ref, head: refHead, export: newClone}
							chs.clonesMu.Lock()
							chs.clones[releaseName] = newCloneForChart
							chs.clonesMu.Unlock()
							if cloneForChart.export != nil {
								cloneForChart.export.Clean()
							}

							// we have a (new) clone, enqueue a release
							cacheKey, err := cache.MetaNamespaceKeyFunc(hr.GetObjectMeta())
							if err != nil {
								continue
							}
							chs.logger.Log("info", "enqueing release upgrade due to change in git chart source", "resource", hr.ResourceID().String())
							chs.releaseQueue.AddRateLimited(cacheKey)
						}
					}
				}
			case <-stopCh:
				chs.logger.Log("stopping", "true")
				return
			}
		}
	}()
}

func mirrorName(chartSource *helmfluxv1.GitChartSource) string {
	return chartSource.GitURL // TODO(michael) this will not always be the case; e.g., per namespace, per auth
}

// maybeMirror starts mirroring the repo needed by a HelmRelease,
// if necessary
func (chs *ChartChangeSync) maybeMirror(hr helmfluxv1.HelmRelease) {
	chartSource := hr.Spec.ChartSource.GitChartSource
	if chartSource != nil {
		if ok := chs.mirrors.Mirror(
			mirrorName(chartSource),
			git.Remote{chartSource.GitURL}, git.Timeout(chs.config.GitTimeout), git.PollInterval(chs.config.GitPollInterval), git.ReadOnly,
		); !ok {
			chs.logger.Log("info", "started mirroring repo", "repo", chartSource.GitURL)
		}
	}
}

// CompareValuesChecksum recalculates the checksum of the values
// and compares it to the last recorded checksum.
func (chs *ChartChangeSync) CompareValuesChecksum(hr helmfluxv1.HelmRelease) bool {
	chartPath, ok := "", false
	if hr.Spec.ChartSource.GitChartSource != nil {
		// We need to hold the lock until have compared the values,
		// so that the clone doesn't get swapped out from under us.
		chs.clonesMu.Lock()
		defer chs.clonesMu.Unlock()
		chartPath, _, ok = chs.getGitChartSource(hr)
		if !ok {
			return false
		}
	} else if hr.Spec.ChartSource.RepoChartSource != nil {
		chartPath, _, ok = chs.getRepoChartSource(hr)
		if !ok {
			return false
		}
	}

	values, err := release.Values(chs.kubeClient.CoreV1(), hr.Namespace, chartPath, hr.GetValuesFromSources(), hr.Spec.Values)
	if err != nil {
		return false
	}

	strValues, err := values.YAML()
	if err != nil {
		return false
	}

	return hr.Status.ValuesChecksum == release.ValuesChecksum([]byte(strValues))
}

// ReconcileReleaseDef asks the ChartChangeSync to examine the release
// associated with a HelmRelease, and install or upgrade the
// release if the chart it refers to has changed.
func (chs *ChartChangeSync) ReconcileReleaseDef(r *release.Release, hr helmfluxv1.HelmRelease) {
	defer chs.updateObservedGeneration(hr)

	releaseName := hr.GetReleaseName()
	logger := log.With(chs.logger, "release", releaseName, "targetNamespace", hr.GetTargetNamespace(), "resource", hr.ResourceID().String())

	// Attempt to retrieve an upgradable release, in case no release
	// or error is returned, install it.
	rel, err := r.GetUpgradableRelease(hr.GetTargetNamespace(), releaseName)
	if err != nil {
		logger.Log("warning", "unable to proceed with release", "err", err)
		return
	}

	opts := release.InstallOptions{DryRun: false}

	chartPath, chartRevision, ok := "", "", false
	if hr.Spec.ChartSource.GitChartSource != nil {
		// We need to hold the lock until after we're done releasing
		// the chart, so that the clone doesn't get swapped out from
		// under us. TODO(michael) consider having a lock per clone.
		chs.clonesMu.Lock()
		defer chs.clonesMu.Unlock()
		chartPath, chartRevision, ok = chs.getGitChartSource(hr)
		if !ok {
			return
		}
		chs.setCondition(hr, helmfluxv1.HelmReleaseChartFetched, v1.ConditionTrue, ReasonCloned, "successfully cloned git repo")
	} else if hr.Spec.ChartSource.RepoChartSource != nil {
		chartPath, chartRevision, ok = chs.getRepoChartSource(hr)
		if !ok {
			return
		}
		chs.setCondition(hr, helmfluxv1.HelmReleaseChartFetched, v1.ConditionTrue, ReasonDownloaded, "chart fetched: "+filepath.Base(chartPath))
	}

	if rel == nil {
		_, checksum, err := r.Install(chartPath, releaseName, hr, release.InstallAction, opts, &chs.kubeClient)
		if err != nil {
			chs.setCondition(hr, helmfluxv1.HelmReleaseReleased, v1.ConditionFalse, ReasonInstallFailed, err.Error())
			chs.logger.Log("warning", "failed to install chart", "err", err)
			return
		}
		chs.setCondition(hr, helmfluxv1.HelmReleaseReleased, v1.ConditionTrue, ReasonSuccess, "helm install succeeded")
		if err = status.SetReleaseRevision(chs.ifClient.HelmV1().HelmReleases(hr.Namespace), hr, chartRevision); err != nil {
			chs.logger.Log("warning", "could not update the release revision", "err", err)
		}
		if err = status.SetValuesChecksum(chs.ifClient.HelmV1().HelmReleases(hr.Namespace), hr, checksum); err != nil {
			chs.logger.Log("warning", "could not update the values checksum", "err", err)
		}
		return
	}

	if !r.ManagedByHelmRelease(rel, hr) {
		msg := fmt.Sprintf("release '%s' does not belong to HelmRelease", releaseName)
		chs.setCondition(hr, helmfluxv1.HelmReleaseReleased, v1.ConditionFalse, ReasonUpgradeFailed, msg)
		chs.logger.Log("warning", msg+", this may be an indication that multiple HelmReleases with the same release name exist")
		return
	}

	changed, err := chs.shouldUpgrade(r, chartPath, rel, hr)
	if err != nil {
		chs.logger.Log("warning", "unable to determine if release has changed", "err", err)
		return
	}
	if changed {
		cHr, err := chs.ifClient.HelmV1().HelmReleases(hr.Namespace).Get(hr.Name, metav1.GetOptions{})
		if err != nil {
			chs.logger.Log("warning", "failed to retrieve HelmRelease scheduled for upgrade", "err", err)
			return
		}
		if diff := cmp.Diff(hr.Spec, cHr.Spec); diff != "" {
			chs.logger.Log("warning", "HelmRelease spec has diverged since we calculated if we should upgrade, skipping upgrade")
			return
		}
		_, checksum, err := r.Install(chartPath, releaseName, hr, release.UpgradeAction, opts, &chs.kubeClient)
		if err != nil {
			chs.setCondition(hr, helmfluxv1.HelmReleaseReleased, v1.ConditionFalse, ReasonUpgradeFailed, err.Error())
			if err = status.SetValuesChecksum(chs.ifClient.HelmV1().HelmReleases(hr.Namespace), hr, checksum); err != nil {
				chs.logger.Log("warning", "could not update the values checksum", "err", err)
			}
			chs.logger.Log("warning", "failed to upgrade chart", "err", err)
			chs.RollbackRelease(r, hr)
			return
		}
		chs.setCondition(hr, helmfluxv1.HelmReleaseReleased, v1.ConditionTrue, ReasonSuccess, "helm upgrade succeeded")
		if err = status.SetReleaseRevision(chs.ifClient.HelmV1().HelmReleases(hr.Namespace), hr, chartRevision); err != nil {
			chs.logger.Log("warning", "could not update the release revision", "err", err)
		}
		if err = status.SetValuesChecksum(chs.ifClient.HelmV1().HelmReleases(hr.Namespace), hr, checksum); err != nil {
			chs.logger.Log("warning", "could not update the values checksum", "err", err)
		}
		return
	}
}

// RollbackRelease rolls back a helm release
func (chs *ChartChangeSync) RollbackRelease(r *release.Release, hr helmfluxv1.HelmRelease) {
	defer chs.updateObservedGeneration(hr)

	if !hr.Spec.Rollback.Enable {
		return
	}

	_, err := r.Rollback(hr)
	if err != nil {
		log.With(
			chs.logger,
			"release", hr.GetReleaseName(), "targetNamespace", hr.GetTargetNamespace(), "resource", hr.ResourceID().String(),
		).Log("warning", "unable to rollback chart release", "err", err)
		chs.setCondition(hr, helmfluxv1.HelmReleaseRolledBack, v1.ConditionFalse, ReasonRollbackFailed, err.Error())
	}
	chs.setCondition(hr, helmfluxv1.HelmReleaseRolledBack, v1.ConditionTrue, ReasonSuccess, "helm rollback succeeded")
}

// DeleteRelease deletes the helm release associated with a
// HelmRelease. This exists mainly so that the operator code can
// call it when it is handling a resource deletion.
func (chs *ChartChangeSync) DeleteRelease(r *release.Release, hr helmfluxv1.HelmRelease) {
	// FIXME(michael): these may need to stop mirroring a repo.
	name := hr.GetReleaseName()
	err := r.Uninstall(hr)
	if err != nil {
		log.With(
			chs.logger,
			"release", hr.GetReleaseName(), "targetNamespace", hr.GetTargetNamespace(), "resource", hr.ResourceID().String(),
		).Log("warning", "chart release not deleted", "err", err)
	}

	// Remove the clone we may have for this HelmRelease
	chs.clonesMu.Lock()
	cloneForChart, ok := chs.clones[name]
	if ok {
		if cloneForChart.export != nil {
			cloneForChart.export.Clean()
		}
		delete(chs.clones, name)
	}
	chs.clonesMu.Unlock()
}

// SyncMirrors instructs all mirrors to refresh from their upstream.
func (chs *ChartChangeSync) SyncMirrors() {
	chs.logger.Log("info", "starting mirror sync")
	for _, err := range chs.mirrors.RefreshAll(chs.config.GitTimeout) {
		chs.logger.Log("error", fmt.Sprintf("failure while syncing mirror: %s", err))
	}
	chs.logger.Log("info", "finished syncing mirrors")
}

// getCustomResourcesForMirror retrieves all the resources that make
// use of the given mirror from the lister.
func (chs *ChartChangeSync) getCustomResourcesForMirror(mirror string) ([]helmfluxv1.HelmRelease, error) {
	var hrs []helmfluxv1.HelmRelease
	list, err := chs.hrLister.List(labels.Everything())
	if err != nil {
		return nil, err
	}

	for _, hr := range list {
		if hr.Spec.GitChartSource == nil {
			continue
		}
		if mirror != mirrorName(hr.Spec.GitChartSource) {
			continue
		}
		hrs = append(hrs, *hr)
	}
	return hrs, nil
}

// setCondition saves the status of a condition.
func (chs *ChartChangeSync) setCondition(hr helmfluxv1.HelmRelease, typ helmfluxv1.HelmReleaseConditionType, st v1.ConditionStatus, reason, message string) error {
	hrClient := chs.ifClient.HelmV1().HelmReleases(hr.Namespace)
	condition := status.NewCondition(typ, st, reason, message)
	return status.SetCondition(hrClient, hr, condition)
}

// updateObservedGeneration updates the observed generation of the
// given HelmRelease to the generation.
func (chs *ChartChangeSync) updateObservedGeneration(hr helmfluxv1.HelmRelease) error {
	hrClient := chs.ifClient.HelmV1().HelmReleases(hr.Namespace)

	return status.SetObservedGeneration(hrClient, hr, hr.Generation)
}

func (chs *ChartChangeSync) getGitChartSource(hr helmfluxv1.HelmRelease) (string, string, bool) {
	chartPath, chartRevision := "", ""
	chartSource := hr.Spec.GitChartSource
	if chartSource == nil {
		return chartPath, chartRevision, false
	}

	releaseName := hr.GetReleaseName()
	logger := log.With(chs.logger, "release", releaseName, "targetNamespace", hr.GetTargetNamespace(), "resource", hr.ResourceID().String())

	chartClone, ok := chs.clones[releaseName]
	// Validate the clone we have for the release is the same as
	// is being referenced in the chart source.
	if ok {
		ok = chartClone.remote == chartSource.GitURL && chartClone.ref == chartSource.RefOrDefault()
		if !ok {
			if chartClone.export != nil {
				chartClone.export.Clean()
			}
			delete(chs.clones, releaseName)
		}
	}

	// FIXME(michael): if it's not cloned, and it's not going to
	// be, we might not want to wait around until the next tick
	// before reporting what's wrong with it. But if we just use
	// repo.Ready(), we'll force all charts through that blocking
	// code, rather than waiting for things to sync in good time.
	if !ok {
		repo, ok := chs.mirrors.Get(mirrorName(chartSource))
		if !ok {
			chs.maybeMirror(hr)
			chs.setCondition(hr, helmfluxv1.HelmReleaseChartFetched, v1.ConditionUnknown, ReasonGitNotReady, "git repo "+chartSource.GitURL+" not mirrored yet")
			logger.Log("info", "chart repo not cloned yet")
		} else {
			status, err := repo.Status()
			if status != git.RepoReady {
				chs.setCondition(hr, helmfluxv1.HelmReleaseChartFetched, v1.ConditionUnknown, ReasonGitNotReady, "git repo not mirrored yet: "+err.Error())
				logger.Log("info", "chart repo not ready yet", "status", string(status), "err", err)
			}
		}
		return chartPath, chartRevision, false
	}
	chartPath = filepath.Join(chartClone.export.Dir(), chartSource.Path)
	chartRevision = chartClone.head

	if chs.config.UpdateDeps && !hr.Spec.ChartSource.GitChartSource.SkipDepUpdate {
		c, ok := chs.helmClients.Load(hr.GetHelmVersion())
		if !ok {
			err := "no Helm client for " + hr.GetHelmVersion()
			chs.setCondition(hr, helmfluxv1.HelmReleaseReleased, v1.ConditionFalse, ReasonDependencyFailed, err)
			logger.Log("warning", "failed to update chart dependencies", "err", err)
			return chartPath, chartRevision, false
		}
		if err := c.DependencyUpdate(chartPath); err != nil {
			chs.setCondition(hr, helmfluxv1.HelmReleaseReleased, v1.ConditionFalse, ReasonDependencyFailed, err.Error())
			logger.Log("warning", "failed to update chart dependencies", "err", err)
			return chartPath, chartRevision, false
		}
	}

	return chartPath, chartRevision, true
}

func (chs *ChartChangeSync) getRepoChartSource(hr helmfluxv1.HelmRelease) (string, string, bool) {
	chartPath, chartRevision := "", ""
	chartSource := hr.Spec.ChartSource.RepoChartSource
	if chartSource == nil {
		return chartPath, chartRevision, false
	}

	path, err := ensureChartFetched(chs.config.ChartCache, chartSource)
	if err != nil {
		chs.setCondition(hr, helmfluxv1.HelmReleaseChartFetched, v1.ConditionFalse, ReasonDownloadFailed, "chart download failed: "+err.Error())
		chs.logger.Log("info", "chart download failed", "resource", hr.ResourceID().String(), "err", err)
		return chartPath, chartRevision, false
	}

	chartPath = path
	chartRevision = chartSource.Version

	return chartPath, chartRevision, true
}

// shouldUpgrade returns true if the current running manifests or chart
// don't match what the repo says we ought to be running, based on
// doing a dry run install from the chart in the git repo.
func (chs *ChartChangeSync) shouldUpgrade(r *release.Release, chartsRepo string, currRel *helm.Release,
	hr helmfluxv1.HelmRelease) (bool, error) {

	if currRel == nil {
		return false, fmt.Errorf("no chart release provided for [%s]", hr.GetName())
	}

	currVals := currRel.Values
	currChart := currRel.Chart

	// Get the desired release state
	opts := release.InstallOptions{DryRun: true}
	tempRelName := string(hr.UID)
	desRel, _, err := r.Install(chartsRepo, tempRelName, hr, release.InstallAction, opts, &chs.kubeClient)
	if err != nil {
		return false, err
	}
	desVals := desRel.Values
	desChart := desRel.Chart

	// compare manifests
	if diff := cmp.Diff(currVals, desVals); diff != "" {
		if chs.config.LogDiffs {
			log.With(
				chs.logger,
				"release", hr.GetReleaseName(), "targetNamespace", hr.GetTargetNamespace(), "resource", hr.ResourceID().String(),
			).Log("info", fmt.Sprintf("release %s: values have diverged", currRel.Name), "diff", diff)
		}
		return true, nil
	}

	// compare chart
	if diff := cmp.Diff(currChart, desChart); diff != "" {
		if chs.config.LogDiffs {
			log.With(
				chs.logger,
				"release", hr.GetReleaseName(), "targetNamespace", hr.GetTargetNamespace(), "resource", hr.ResourceID().String(),
			).Log("info", fmt.Sprintf("release %s: chart has diverged", currRel.Name), "resource", hr.ResourceID().String(), "diff", diff)
		}
		return true, nil
	}

	return false, nil
}
