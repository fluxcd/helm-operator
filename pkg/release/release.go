package release

import (
	"errors"
	"path/filepath"
	"time"

	"github.com/fluxcd/flux/pkg/git"
	"github.com/go-kit/kit/log"
	"github.com/google/go-cmp/cmp"

	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/fluxcd/helm-operator/pkg/annotator"
	"github.com/fluxcd/helm-operator/pkg/apis/helm.fluxcd.io/v1"
	"github.com/fluxcd/helm-operator/pkg/chartsync"
	v1client "github.com/fluxcd/helm-operator/pkg/client/clientset/versioned/typed/helm.fluxcd.io/v1"
	"github.com/fluxcd/helm-operator/pkg/helm"
	"github.com/fluxcd/helm-operator/pkg/status"
)

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
	annotator         *annotator.Annotator
	config            Config
}

func New(logger log.Logger, coreV1Client corev1client.CoreV1Interface, helmReleaseClient v1client.HelmV1Interface,
	gitChartSync *chartsync.GitChartSync, annotator *annotator.Annotator, config Config) *Release {

	r := &Release{
		logger:            logger,
		coreV1Client:      coreV1Client,
		helmReleaseClient: helmReleaseClient,
		gitChartSync:      gitChartSync,
		annotator:         annotator,
		config:            config.WithDefaults(),
	}
	return r
}

// Sync synchronizes the given `v1.HelmRelease` with Helm.
func (r *Release) Sync(client helm.Client, hr *v1.HelmRelease) (rHr *v1.HelmRelease, err error) {
	defer func(start time.Time) {
		ObserveRelease(start, err == nil, hr.GetTargetNamespace(), hr.GetReleaseName())
	}(time.Now())
	defer status.SetObservedGeneration(r.helmReleaseClient.HelmReleases(hr.Namespace), hr, hr.Generation)

	logger := releaseLogger(r.logger, client, hr)

	// Ensure we have the chart for the release, construct the path
	// to the chart, and record the revision.
	var chartPath, revision string
	switch {
	case hr.Spec.GitChartSource != nil && hr.Spec.GitURL != "" && hr.Spec.Path != "":
		var export *git.Export
		var err error

		export, revision, err = r.gitChartSync.GetMirrorCopy(hr)
		if err != nil {
			switch err.(type) {
			case chartsync.ChartUnavailableError:
				_ = status.SetCondition(r.helmReleaseClient.HelmReleases(hr.Namespace), hr, status.NewCondition(
					v1.HelmReleaseChartFetched, v1.ConditionFalse, ReasonDownloadFailed, err.Error()))
			case chartsync.ChartNotReadyError:
				_ = status.SetCondition(r.helmReleaseClient.HelmReleases(hr.Namespace), hr, status.NewCondition(
					v1.HelmReleaseChartFetched, v1.ConditionUnknown, ReasonGitNotReady, err.Error()))
			}
			logger.Log("error", err.Error())
			return hr, err
		}

		defer export.Clean()
		chartPath = filepath.Join(export.Dir(), hr.Spec.GitChartSource.Path)

		_ = status.SetCondition(r.helmReleaseClient.HelmReleases(hr.Namespace), hr, status.NewCondition(
			v1.HelmReleaseChartFetched, v1.ConditionTrue, ReasonGitCloned, "successfully cloned chart revision: "+revision))

		if r.config.UpdateDeps && !hr.Spec.GitChartSource.SkipDepUpdate {
			// Attempt to update chart dependencies, if it fails we
			// simply update the status on the resource and return.
			if err := client.DependencyUpdate(chartPath); err != nil {
				_ = status.SetCondition(r.helmReleaseClient.HelmReleases(hr.Namespace), hr, status.NewCondition(
					v1.HelmReleaseReleased, v1.ConditionFalse, ReasonDependencyFailed, err.Error()))
				logger.Log("error", ErrDepUpdate.Error(), "err", err.Error())
				return hr, err
			}
		}
	case hr.Spec.RepoChartSource != nil:
		var fetched bool
		var err error

		chartPath, fetched, err = chartsync.EnsureChartFetched(client, r.config.ChartCache, hr.Spec.RepoChartSource)
		revision = hr.Spec.RepoChartSource.Version

		if err != nil {
			_ = status.SetCondition(r.helmReleaseClient.HelmReleases(hr.Namespace), hr, status.NewCondition(
				v1.HelmReleaseChartFetched, v1.ConditionFalse, ReasonDownloadFailed, err.Error()))
			logger.Log("error", err.Error())
			return hr, err
		}
		if fetched {
			_ = status.SetCondition(r.helmReleaseClient.HelmReleases(hr.Namespace), hr, status.NewCondition(
				v1.HelmReleaseChartFetched, v1.ConditionTrue, ReasonDownloaded, "chart fetched: "+filepath.Base(chartPath)))
		}
	default:
		_ = status.SetCondition(r.helmReleaseClient.HelmReleases(hr.Namespace), hr, status.NewCondition(
			v1.HelmReleaseChartFetched, v1.ConditionFalse, ReasonDownloadFailed, ErrNoChartSource.Error()))
		logger.Log("error", ErrNoChartSource.Error())
		return hr, ErrNoChartSource
	}

	// Check if a release already exists, this is used to determine
	// if and how we should sync, and what actions we should take
	// if the sync fails.
	curRel, err := client.Get(hr.GetReleaseName(), helm.GetOptions{Namespace: hr.GetTargetNamespace()})
	if err != nil {
		_ = status.SetCondition(r.helmReleaseClient.HelmReleases(hr.Namespace), hr, status.NewCondition(
			v1.HelmReleaseReleased, v1.ConditionFalse, ReasonClientError, err.Error()))
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
		_ = status.SetCondition(r.helmReleaseClient.HelmReleases(hr.Namespace), hr, status.NewCondition(
			v1.HelmReleaseReleased, v1.ConditionFalse, failReason, ErrComposingValues.Error()))
		logger.Log("error", ErrComposingValues.Error(), "err", err.Error())
		return hr, ErrComposingValues
	}

	ok, err := shouldSync(logger, r.annotator, client, hr, curRel, chartPath, composedValues, r.config.LogDiffs)
	if !ok {
		if err != nil {
			_ = status.SetCondition(r.helmReleaseClient.HelmReleases(hr.Namespace), hr, status.NewCondition(
				v1.HelmReleaseReleased, v1.ConditionFalse, failReason, err.Error()))
			logger.Log("error", ErrShouldSync.Error(), "err", err.Error())
			return hr, ErrShouldSync
		}
		return hr, nil
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
		SkipCRDs:    hr.Spec.SkipCRDs,
		MaxHistory:  hr.GetMaxHistory(),
		Wait:        hr.Spec.Wait || (curRel != nil && hr.Spec.Rollback.Enable),
	})
	if err != nil {
		_ = status.SetCondition(r.helmReleaseClient.HelmReleases(hr.Namespace), hr, status.NewCondition(
			v1.HelmReleaseReleased, v1.ConditionFalse, failReason, err.Error()))
		logger.Log("error", "Helm release failed", "revision", revision, "err", err.Error())

		// If there was no release prior to this,
		// uninstall the failed release so it can be retried.
		if curRel == nil {
			logger.Log("info", "uninstalling initial failed release so it can be retried")
			if uErr := client.Uninstall(hr.GetReleaseName(), helm.UninstallOptions{
				Namespace:   hr.GetTargetNamespace(),
				KeepHistory: false,
				Timeout:     hr.GetTimeout(),
			}); uErr != nil {
				logger.Log("error", "Helm uninstall failed", "err", uErr.Error())
			}
			return hr, err
		}

		// Rollbacks are not enabled; return and wait for the next signal to retry...
		if !hr.Spec.Rollback.Enable {
			return hr, err
		}

		// Determine if a release actually happened, as with Helm v3
		// it is possible an i.e. validation error was returned while
		// attempting to make a release, rolling back on this would
		// either result in going back to a wrong version, or the
		// complete removal of the Helm release.
		//
		// TODO(hidde): it would be better if we were able to act on
		// the returned error. Doing this would however mean that we
		// need to be able to match the errors with certainty, which
		// is currently not possible as all returned errors are
		// flattened and 'type checking' is thus only possible by
		// performing string matches; a fairly insecure operation.
		// With a bit of luck the upstream libraries will eventually
		// move to the '%w' error wrapping added in Golang 1.13,
		// making all of this a lot easier.
		newRel, rErr := client.Get(hr.GetReleaseName(), helm.GetOptions{Namespace: hr.GetTargetNamespace()})
		if rErr != nil {
			logger.Log("error", "failed to determine if Helm release can be rolled back", "err", err.Error())
			return hr, rErr
		}
		if newRel.Version != (curRel.Version + 1) {
			return hr, err
		}

		performRollback = true
	} else {
		_ = status.SetCondition(r.helmReleaseClient.HelmReleases(hr.Namespace), hr, status.NewCondition(
			v1.HelmReleaseReleased, v1.ConditionTrue, ReasonSuccess, "Helm release sync succeeded"))
		status.SetReleaseRevision(r.helmReleaseClient.HelmReleases(hr.Namespace), hr, revision)
		logger.Log("info", "Helm release sync succeeded", "revision", revision)
	}

	// The upgrade attempt failed, rollback if instructed...
	if performRollback {
		logger.Log("info", "rolling back failed Helm release")
		rel, err = client.Rollback(hr.GetReleaseName(), helm.RollbackOptions{
			Namespace:    hr.GetTargetNamespace(),
			Timeout:      hr.Spec.Rollback.GetTimeout(),
			Wait:         hr.Spec.Rollback.Wait,
			DisableHooks: hr.Spec.Rollback.DisableHooks,
			Recreate:     hr.Spec.Rollback.Recreate,
			Force:        hr.Spec.Rollback.Force,
		})
		if err != nil {
			_ = status.SetCondition(r.helmReleaseClient.HelmReleases(hr.Namespace), hr, status.NewCondition(
				v1.HelmReleaseRolledBack, v1.ConditionFalse, ReasonRollbackFailed, err.Error()))
			logger.Log("error", "Helm rollback failed", "err", err.Error())
			return hr, err
		}
		_ = status.SetCondition(r.helmReleaseClient.HelmReleases(hr.Namespace), hr, status.NewCondition(
			v1.HelmReleaseRolledBack, v1.ConditionTrue, ReasonSuccess, "Helm rollback succeeded"))
		logger.Log("info", "Helm rollback succeeded")

		// We should still report failure.
		err = ErrRolledBack
	}

	if err := r.annotator.Annotate(rel.Resources, rel.Namespace, v1.AntecedentAnnotation, hr.ResourceID().String()); err != nil {
		r.logger.Log("error", "failed to annotate some resources", "err", err.Error())
	}

	return hr, err
}

// Uninstalls removes the Helm release for the given `v1.HelmRelease`,
// and the git chart source if present.
func (r *Release) Uninstall(client helm.Client, hr *v1.HelmRelease) {
	logger := releaseLogger(r.logger, client, hr)

	if err := client.Uninstall(hr.GetReleaseName(), helm.UninstallOptions{
		Namespace:   hr.GetTargetNamespace(),
		KeepHistory: false,
		Timeout:     hr.GetTimeout(),
	}); err != nil {
		logger.Log("error", "failed to uninstall Helm release", "err", err.Error())
	}

	if hr.Spec.GitChartSource != nil {
		r.gitChartSync.Delete(hr)
	}
}

// shouldSync determines if the given `v1.HelmRelease` should be synced
// with Helm. The cheapest checks which do not require a dry-run are
// consulted first (e.g. is this our first sync, have we already seen
// this revision of the resource); before running the dry-run release to
// determine if any undefined mutations have occurred. It returns a
// booleans indicating if the release should be synced, or an error.
func shouldSync(logger log.Logger, annotator *annotator.Annotator, client helm.Client, hr *v1.HelmRelease, curRel *helm.Release,
	chartPath string, values helm.Values, logDiffs bool) (bool, error) {

	// Without valid YAML we will not get anywhere, return early.
	b, err := values.YAML()
	if err != nil {
		return false, ErrComposingValues
	}

	// If there is no existing release, we should simply sync.
	if curRel == nil {
		logger.Log("info", "no existing release", "action", "install")
		return true, nil
	}

	// Check if the release is managed by our resource
	managedBy, resourceID, err := annotator.OneHasAnnotationWithValueOrNil(curRel.Resources, curRel.Namespace, v1.AntecedentAnnotation, hr.ResourceID().String())
	// If an error is returned and we were unable to determine ownership,
	// we return the error (and skip) to avoid conflicts.
	if err != nil {
		return false, err
	}
	// If the release is not managed by our resource, we skip to avoid conflicts.
	if !managedBy {
		logger.Log("warning", "release appears to be managed by "+resourceID, "action", "skip")
		return false, nil
	}

	// If the current state of the release does not allow us to safely upgrade, we skip.
	if s := curRel.Info.Status; !s.AllowsUpgrade() {
		logger.Log("warning", "unable to sync release with status "+s.String(), "action", "skip")
		return false, nil
	}

	// If we have not processed this generation of the release, we should sync.
	if !status.HasSynced(*hr) {
		logger.Log("info", "release has not yet been processed", "action", "upgrade")
		return true, nil
	}

	// Next, we perform a dry-run upgrade and compare the result against the
	// latest release _or_ the latest failed release in case of a rollback.
	// If this results in one or more diffs we should sync.
	logger.Log("info", "performing dry-run upgrade to see if release has diverged")
	desRel, err := client.UpgradeFromPath(chartPath, hr.GetReleaseName(), b, helm.UpgradeOptions{
		DryRun:      true,
		Namespace:   hr.GetTargetNamespace(),
		Force:       hr.Spec.ForceUpgrade,
		ResetValues: hr.Spec.ResetValues,
	})
	if err != nil {
		return false, err
	}

	var vDiff, cDiff string
	switch {
	case status.HasRolledBack(*hr):
		if status.ShouldRetryUpgrade(*hr) {
			logger.Log("info", "release has been rolled back", "rollbackCount", hr.Status.RollbackCount, "maxRetries", hr.Spec.Rollback.GetMaxRetries(), "action", "upgrade")
			return true, nil
		}
		logger.Log("info", "release has been rolled back, comparing dry-run output with latest failed release")
		rels, err := client.History(hr.GetReleaseName(), helm.HistoryOptions{Namespace: hr.GetTargetNamespace()})
		if err != nil {
			return false, err
		}
		for _, r := range rels {
			if r.Info.Status == helm.StatusFailed {
				vDiff, cDiff = compareRelease(r, desRel)
				break
			}
		}
	default:
		vDiff, cDiff = compareRelease(curRel, desRel)
	}

	if vDiff != "" && logDiffs {
		logger.Log("info", "values have diverged", "diff", vDiff)
	}

	if cDiff != "" && logDiffs {
		logger.Log("info", "chart has diverged", "diff", cDiff)
	}

	if cDiff != "" || vDiff != "" {
		logger.Log("info", "dry-run differed", "action", "upgrade")
		return true, nil
	}

	logger.Log("info", "no changes", "action", "skip")
	return false, nil
}

// compareRelease compares the values and charts of the two given
// releases and returns the diff sets.
func compareRelease(j *helm.Release, k *helm.Release) (string, string) {
	return cmp.Diff(j.Values, k.Values), cmp.Diff(j.Chart, k.Chart)
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
