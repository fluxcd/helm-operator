package release

import (
	"errors"
	"path/filepath"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/google/go-cmp/cmp"

	corev1 "k8s.io/api/core/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"

	v1 "github.com/fluxcd/helm-operator/pkg/apis/helm.fluxcd.io/v1"
	"github.com/fluxcd/helm-operator/pkg/chartsync"
	v1client "github.com/fluxcd/helm-operator/pkg/client/clientset/versioned/typed/helm.fluxcd.io/v1"
	"github.com/fluxcd/helm-operator/pkg/helm"
	"github.com/fluxcd/helm-operator/pkg/status"
)

const maxHistory = 10

// Condition change reasons.
const (
	ReasonGitNotReady      = "GitNotReady"
	ReasonGitCloned        = "GitRepoCloned"
	ReasonDownloadFailed   = "RepoFetchFailed"
	ReasonDownloaded       = "RepoChartInCache"
	ReasonInstallFailed    = "HelmInstallFailed"
	ReasonClientError      = "HelmClientError"
	ReasonDependencyFailed = "UpdateDependencyFailed"
	ReasonUpgradeFailed    = "HelmUpgradeFailed"
	ReasonRollbackFailed   = "HelmRollbackFailed"
	ReasonSuccess          = "HelmSuccess"
)

// Various (final) errors.
var (
	ErrDepUpdate       = errors.New("failed updating dependencies")
	ErrNoChartSource   = errors.New("no chart source given")
	ErrComposingValues = errors.New("failed to compose values for chart release")
	ErrShouldSync      = errors.New("failed to determine if the release should be synced")
	ErrRolledBack      = errors.New("upgrade failed and release has been rolled back")
)

// Config holds the configuration for releases.
type Config struct {
	ChartCache string
	UpdateDeps bool
	LogDiffs   bool
}

// WithDefaults sets the default values for the release config.
func (c Config) WithDefaults() Config {
	if c.ChartCache == "" {
		c.ChartCache = "/tmp"
	}
	return c
}

// Release holds the elements required to perform a Helm release,
// and provides the methods to perform a sync or uninstall.
type Release struct {
	logger            log.Logger
	coreV1Client      corev1client.CoreV1Interface
	helmReleaseClient v1client.HelmV1Interface
	gitChartSync      *chartsync.GitChartSync
	config            Config
}

func New(logger log.Logger, coreV1Client corev1client.CoreV1Interface, helmReleaseClient v1client.HelmV1Interface,
	gitChartSync *chartsync.GitChartSync, config Config) *Release {

	r := &Release{
		logger:            logger,
		coreV1Client:      coreV1Client,
		helmReleaseClient: helmReleaseClient,
		gitChartSync:      gitChartSync,
		config:            config.WithDefaults(),
	}
	return r
}

// Sync synchronizes the given `v1.HelmRelease` with Helm.
func (r *Release) Sync(client helm.Client, hr *v1.HelmRelease) (rHr *v1.HelmRelease, err error) {
	defer func(start time.Time) {
		ObserveRelease(start, err == nil, hr.GetTargetNamespace(), hr.GetReleaseName())
	}(time.Now())
	defer status.SetObservedGeneration(r.helmReleaseClient.HelmReleases(hr.Namespace), *hr, hr.Generation)

	logger := releaseLogger(r.logger, client, hr)

	// Ensure we have the chart for the release, construct the path
	// to the chart, and record the revision.
	var chartPath, revision string
	switch {
	case hr.Spec.GitChartSource != nil:
		export, revision, err := r.gitChartSync.GetMirrorCopy(hr)
		if err != nil {
			switch err.(type) {
			case chartsync.ChartUnavailableError:
				_ = status.SetCondition(r.helmReleaseClient.HelmReleases(hr.Namespace), *hr, status.NewCondition(
					v1.HelmReleaseChartFetched, corev1.ConditionFalse, ReasonDownloadFailed, err.Error()))
			case chartsync.ChartNotReadyError:
				_ = status.SetCondition(r.helmReleaseClient.HelmReleases(hr.Namespace), *hr, status.NewCondition(
					v1.HelmReleaseChartFetched, corev1.ConditionUnknown, ReasonGitNotReady, err.Error()))
			}
			logger.Log("error", err.Error())
			return hr, err
		}

		defer export.Clean()
		chartPath = filepath.Join(export.Dir(), hr.Spec.GitChartSource.Path)

		_ = status.SetCondition(r.helmReleaseClient.HelmReleases(hr.Namespace), *hr, status.NewCondition(
			v1.HelmReleaseChartFetched, corev1.ConditionTrue, ReasonGitCloned, "successfully cloned chart revision: "+revision))

		if r.config.UpdateDeps && !hr.Spec.GitChartSource.SkipDepUpdate {
			// Attempt to update chart dependencies, if it fails we
			// simply update the status on the resource and return.
			if err := client.DependencyUpdate(chartPath); err != nil {
				_ = status.SetCondition(r.helmReleaseClient.HelmReleases(hr.Namespace), *hr, status.NewCondition(
					v1.HelmReleaseReleased, corev1.ConditionFalse, ReasonDependencyFailed, err.Error()))
				logger.Log("error", ErrDepUpdate.Error(), "err", err.Error())
				return hr, err
			}
		}
	case hr.Spec.RepoChartSource != nil:
		var err error

		chartPath, err = chartsync.EnsureChartFetched(client, r.config.ChartCache, hr.Spec.RepoChartSource)
		revision = hr.Spec.RepoChartSource.Version

		if err != nil {
			_ = status.SetCondition(r.helmReleaseClient.HelmReleases(hr.Namespace), *hr, status.NewCondition(
				v1.HelmReleaseChartFetched, corev1.ConditionFalse, ReasonDownloadFailed, err.Error()))
			logger.Log("error", err.Error())
			return hr, err
		}

		_ = status.SetCondition(r.helmReleaseClient.HelmReleases(hr.Namespace), *hr, status.NewCondition(
			v1.HelmReleaseChartFetched, corev1.ConditionTrue, ReasonDownloaded, "chart fetched: "+filepath.Base(chartPath)))
	default:
		_ = status.SetCondition(r.helmReleaseClient.HelmReleases(hr.Namespace), *hr, status.NewCondition(
			v1.HelmReleaseChartFetched, corev1.ConditionFalse, ReasonDownloadFailed, ErrNoChartSource.Error()))
		logger.Log("error", ErrNoChartSource.Error())
		return hr, ErrNoChartSource
	}

	// Check if a release already exists, this is used to determine
	// if and how we should sync, and what actions we should take
	// if the sync fails.
	curRel, err := client.Status(hr.GetReleaseName(), helm.StatusOptions{Namespace: hr.GetTargetNamespace()})
	if err != nil {
		_ = status.SetCondition(r.helmReleaseClient.HelmReleases(hr.Namespace), *hr, status.NewCondition(
			v1.HelmReleaseReleased, corev1.ConditionFalse, ReasonClientError, err.Error()))
		logger.Log("error", ErrShouldSync.Error(), "err", err.Error())
		return hr, ErrShouldSync
	}

	// Record failure reason for further condition updates.
	failReason := ReasonInstallFailed
	if curRel != nil {
		failReason = ReasonUpgradeFailed
	}

	// Compose the values from the sources and values defined in the
	// `v1.HelmRelease` resource.
	composedValues, err := composeValues(r.coreV1Client, hr, chartPath)
	if err != nil {
		_ = status.SetCondition(r.helmReleaseClient.HelmReleases(hr.Namespace), *hr, status.NewCondition(
			v1.HelmReleaseReleased, corev1.ConditionFalse, failReason, ErrComposingValues.Error()))
		logger.Log("error", ErrComposingValues.Error(), "err", err.Error())
		return hr, ErrComposingValues
	}
	defer status.SetValuesChecksum(r.helmReleaseClient.HelmReleases(hr.Namespace), *hr, composedValues.Checksum())

	if ok, err := shouldSync(logger, client, hr, curRel, chartPath, composedValues, r.config.LogDiffs); !ok {
		if err != nil {
			_ = status.SetCondition(r.helmReleaseClient.HelmReleases(hr.Namespace), *hr, status.NewCondition(
				v1.HelmReleaseReleased, corev1.ConditionFalse, failReason, ErrComposingValues.Error()))
			logger.Log("error", ErrShouldSync.Error(), "err", err.Error())
		}
		return hr, ErrShouldSync
	}

	// `shouldSync` above has already validated the YAML output of our
	// composed values, so we ignore the fact that this could
	// technically return an error.
	v, _ := composedValues.YAML()

	var performRollback bool

	// Off we go! Attempt to perform the actual upgrade.
	rel, err := client.UpgradeFromPath(chartPath, hr.GetReleaseName(), v, helm.UpgradeOptions{
		Namespace:   hr.GetTargetNamespace(),
		Timeout:     hr.GetTimeout(),
		Install:     curRel == nil,
		Force:       hr.Spec.ForceUpgrade,
		ResetValues: hr.Spec.ResetValues,
		MaxHistory:  maxHistory,
		// We only set this during installation to delete a failed
		// release, but not during upgrades, as we ourselves want
		// to be in control of rollbacks.
		Atomic: curRel == nil,
		Wait:   hr.Spec.Rollback.Enable,
	})
	if err != nil {
		_ = status.SetCondition(r.helmReleaseClient.HelmReleases(hr.Namespace), *hr, status.NewCondition(
			v1.HelmReleaseReleased, corev1.ConditionFalse, failReason, err.Error()))
		logger.Log("error", "Helm release failed", "revision", revision, "err", err.Error())

		// If this is the first release, or rollbacks are not enabled;
		// return and wait for the next signal to retry...
		if curRel == nil || !hr.Spec.Rollback.Enable {
			return hr, err
		}

		performRollback = true
	} else {
		_ = status.SetCondition(r.helmReleaseClient.HelmReleases(hr.Namespace), *hr, status.NewCondition(
			v1.HelmReleaseReleased, corev1.ConditionTrue, ReasonSuccess, "Helm release sync succeeded"))
		status.SetReleaseRevision(r.helmReleaseClient.HelmReleases(hr.Namespace), *hr, revision)
		logger.Log("info", "Helm release sync succeeded", "revision", revision)
	}

	// The upgrade attempt failed, rollback if instructed...
	if performRollback {
		rel, err = client.Rollback(hr.GetReleaseName(), helm.RollbackOptions{
			Namespace: hr.GetTargetNamespace(),
			Timeout:   hr.GetTimeout(),
			Force:     hr.Spec.ForceUpgrade,
		})
		if err != nil {
			_ = status.SetCondition(r.helmReleaseClient.HelmReleases(hr.Namespace), *hr, status.NewCondition(
				v1.HelmReleaseRolledBack, corev1.ConditionFalse, ReasonRollbackFailed, err.Error()))
			logger.Log("error", "Helm rollback failed", "err", err.Error())
			return hr, err
		}
		_ = status.SetCondition(r.helmReleaseClient.HelmReleases(hr.Namespace), *hr, status.NewCondition(
			v1.HelmReleaseRolledBack, corev1.ConditionTrue, ReasonSuccess, "Helm rollback succeeded"))
		logger.Log("info", "Helm rollback succeeded")

		// We should still report failure.
		err = ErrRolledBack
	}

	annotateResources(logger, rel, hr.ResourceID())

	return hr, err
}

// Uninstalls removes the Helm release for the given `v1.HelmRelease`,
// and the git chart source if present.
func (r *Release) Uninstall(client helm.Client, hr *v1.HelmRelease) {
	logger := releaseLogger(r.logger, client, hr)

	if err := client.Uninstall(hr.GetReleaseName(), helm.UninstallOptions{
		Namespace:   hr.GetTargetNamespace(),
		KeepHistory: false,
	}); err != nil {
		logger.Log("error", "failed to uninstall Helm release", "err", err.Error())
	}

	if hr.Spec.GitChartSource != nil {
		r.gitChartSync.Delete(hr)
	}
}

// shouldSync determines if the given `v1.HelmRelease` should be synced
// with Helm. The cheapest checks which do not require a dry-run are
// consulted first (e.g. is this our first sync, has the release been
// rolled back, have we already seen this revision of the resource);
// before running the dry-run release to determine if any undefined
// mutations have occurred.
func shouldSync(logger log.Logger, client helm.Client, hr *v1.HelmRelease, curRel *helm.Release,
	chartPath string, values values, logDiffs bool) (bool, error) {

	if curRel == nil {
		logger.Log("info", "no existing release; installing")
		// If there is no existing release, we should simply sync.
		return true, nil
	}

	if ok, resourceID := managedByHelmRelease(curRel, *hr); !ok {
		logger.Log("warning", "release appears to be managed by "+resourceID+"; skipping")
		return false, nil
	}

	if s := curRel.Info.Status; !s.Syncable() {
		logger.Log("warning", "unable to sync release with status "+s.String()+"; skipping")
		return false, nil
	}

	if status.HasRolledBack(*hr) {
		if hr.Status.ValuesChecksum != values.Checksum() {
			// The release has been rolled back but the values have
			// changed. We should attempt a new sync to see if the
			// change resolved the issue that triggered the rollback.
			logger.Log("info", "values appear to have changed since rollback; attempting upgrade")
			return true, nil
		}
		logger.Log("warning", "release has been rolled back; skipping")
		return false, nil
	}

	if !status.HasSynced(*hr) {
		// The generation of this `v1.HelmRelease` has not been
		// processed, we should simply sync.
		return true, nil
	}

	// We use the UID as dry-run release name, as this value is unique,
	// causes no collision with releases that may exist, and does not
	// exceed the max release name length of 53 characters.
	dryRunRelName := string(hr.UID)
	b, err := values.YAML()
	if err != nil {
		// Without valid YAML values we are unable to sync.
		return false, ErrComposingValues
	}

	// Perform the dry-run so that we can compare what we ought to be
	// running matches what is defined in the `v1.HelmRelease`.
	desRel, err := client.UpgradeFromPath(chartPath, dryRunRelName, b, helm.UpgradeOptions{ClientOnly: true, Install: true, DryRun: true})
	if err != nil {
		return false, err
	}

	curValues, desValues := curRel.Values, desRel.Values
	curChart, desChart := curRel.Chart, desRel.Chart

	// Compare values to detect mutations.
	vDiff := cmp.Diff(curValues, desValues)
	if vDiff != "" && logDiffs {
		logger.Log("info", "values have diverged", "diff", vDiff)
	}

	// Compare chart to detect mutations.
	cDiff := cmp.Diff(curChart, desChart)
	if cDiff != "" && logDiffs {
		logger.Log("info", "chart has diverged", "diff", cDiff)
	}

	return vDiff != "" || cDiff != "", nil
}

// releaseLogger returns a logger in the context of the given
// HelmRelease (that being, with metadata included).
func releaseLogger(logger log.Logger, client helm.Client, hr *v1.HelmRelease) log.Logger {
	return log.With(logger,
		"release", hr.GetReleaseName(),
		"targetNamespace", hr.GetTargetNamespace(),
		"resource", hr.ResourceID().String(),
		"helmVersion", client.Version(),
	)
}
