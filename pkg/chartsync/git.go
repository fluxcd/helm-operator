package chartsync

import (
	"context"
	"path/filepath"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/weaveworks/flux/git"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"

	"github.com/fluxcd/helm-operator/pkg/apis/helm.fluxcd.io/v1"
	clientset "github.com/fluxcd/helm-operator/pkg/client/clientset/versioned"
	lister "github.com/fluxcd/helm-operator/pkg/client/listers/helm.fluxcd.io/v1"
	"github.com/fluxcd/helm-operator/pkg/status"
)

// Condition change reasons
const (
	ReasonGitNotReady = "GitNotReady"
	ReasonGitCloned   = "GitRepoCloned"
)

// ReleaseQueue is an add-only `workqueue.RateLimitingInterface`
type ReleaseQueue interface {
	AddRateLimited(item interface{})
}

// GitConfig holds the configuration for git operations.
type GitConfig struct {
	GitTimeout      time.Duration
	GitPollInterval time.Duration
}

// GitChartSource holds the source references to a Helm chart in git,
// and a minimal working clone.
type GitChartSource struct {
	sync.Mutex
	Export *git.Export
	Mirror string
	Remote string
	Ref    string
	Head   string
}

// ChartPath returns the absolute chart path for the given relative
// path.
func (c *GitChartSource) ChartPath(relativePath string) string {
	return filepath.Join(c.Export.Dir(), relativePath)
}

// forHelmRelease returns true if the given `v1.HelmRelease`s
// `v1.GitChartSource` matches the GitChartSource.
func (c *GitChartSource) forHelmRelease(hr *v1.HelmRelease) bool {
	if hr == nil || hr.Spec.GitChartSource == nil {
		return false
	}
	return c.Mirror == mirrorName(hr) && c.Remote == hr.Spec.GitURL && c.Ref == hr.Spec.Ref
}

// GitChartSourceSync syncs `GitChartSource`s with their mirrors.
type GitChartSourceSync struct {
	logger log.Logger
	config GitConfig

	lister lister.HelmReleaseLister
	client clientset.Interface

	mirrors *git.Mirrors

	sourcesMu sync.Mutex
	sources   map[string]*GitChartSource

	releaseQueue ReleaseQueue
}

func NewGitChartSourceSync(logger log.Logger,
	lister lister.HelmReleaseLister, cfg GitConfig, queue ReleaseQueue) *GitChartSourceSync {

	return &GitChartSourceSync{
		logger:       logger,
		config:       cfg,
		lister:       lister,
		mirrors:      git.NewMirrors(),
		sourcesMu:    sync.Mutex{},
		sources:      make(map[string]*GitChartSource),
		releaseQueue: queue,
	}
}

// Run starts the sync of `GitChartSource`s.
func (c *GitChartSourceSync) Run(stopCh <-chan struct{}, errCh chan error, wg *sync.WaitGroup) {
	c.logger.Log("info", "starting sync of git chart sources")

	wg.Add(1)
	go func() {
		defer runtime.HandleCrash()
		defer func() {
			c.mirrors.StopAllAndWait()
			wg.Done()
		}()

		for {
			select {
			case changed := <-c.mirrors.Changes():
				for mirrorName := range changed {
					mirror, ok := c.mirrors.Get(mirrorName)

					// Get the HelmReleases that make use of the mirror.
					hrs, err := c.getHelmReleasesForMirror(mirrorName)
					if err != nil {
						c.logger.Log("error", "failed to get git chart sources for mirror", "mirror", mirrorName, "err", err)
						continue
					}

					// We received a signal from a no longer existing
					// mirror.
					if !ok {
						if len(hrs) == 0 {
							// If there are no references to it either,
							// just continue with the next mirror...
							continue
						}

						c.logger.Log("warning", "no existing mirror found for signaled git change", "mirror", mirrorName)
						for _, hr := range hrs {
							nsClient := c.client.HelmV1().HelmReleases(hr.Namespace)
							_ = status.SetCondition(nsClient, *hr, status.NewCondition(
								v1.HelmReleaseChartFetched,
								corev1.ConditionUnknown,
								ReasonGitNotReady,
								"git mirror missing; starting mirroring again",
							))
							c.maybeMirror(mirrorName, hr.Spec.GitChartSource.GitURL)
						}
						// Wait for the signal from the newly requested mirror...
						continue
					}
					if mirrorStatus, err := mirror.Status(); mirrorStatus != git.RepoReady {
						c.logger.Log("warning", "mirror not ready for sync", "status", mirrorStatus, "mirror", mirrorName)
						for _, hr := range hrs {
							nsClient := c.client.HelmV1().HelmReleases(hr.Namespace)
							_ = status.SetCondition(nsClient, *hr, status.NewCondition(
								v1.HelmReleaseChartFetched,
								corev1.ConditionUnknown,
								ReasonGitNotReady,
								err.Error(),
							))
						}
						// Wait for the next signal...
						continue
					}
					for _, hr := range hrs {
						if _, ok := c.syncGitChartSource(mirror, hr); ok {
							cacheKey, err := cache.MetaNamespaceKeyFunc(hr.GetObjectMeta())
							if err != nil {
								continue // this should never happen
							}
							c.releaseQueue.AddRateLimited(cacheKey)
						}
					}
				}
			case <-stopCh:
				c.logger.Log("info", "stopping sync of git chart sources")
				return
			}
		}
	}()
}

// Load returns a pointer to the requested `GitChartSource` for the given
// `v1.HelmRelease` if found, and a boolean indicating success.
func (c *GitChartSourceSync) Load(hr *v1.HelmRelease) (*GitChartSource, bool) {

	// Check if we have a source in store and return if it still equals
	// to what is configured in the source.
	c.sourcesMu.Lock()
	cc, ok := c.sources[hr.ResourceID().String()]
	c.sourcesMu.Unlock()
	if ok && cc.forHelmRelease(hr) {
		return cc, true
	}

	// Check if there is an existing mirror .
	mirror := mirrorName(hr)
	repo, ok := c.mirrors.Get(mirror)
	if !ok {
		// We did not find a mirror; request one, return, and wait for
		// signal.
		c.maybeMirror(mirror, hr.Spec.GitURL)
		return nil, false
	}

	if s, _ := c.syncGitChartSource(repo, hr); s != nil {
		return s, true
	}
	return nil, false
}

// Delete cleans up the git chart source for the given resource ID,
// this includes the mirror if there is no reference to it from sources.
// It returns a boolean indicating a successful removal (`true` if so,
// `false` otherwise).
func (c *GitChartSourceSync) Delete(hr *v1.HelmRelease) bool {
	c.sourcesMu.Lock()
	defer c.sourcesMu.Unlock()

	// Attempt to get the source from store.
	source, ok := c.sources[hr.ResourceID().String()]
	if ok {
		source.Lock()
		if source.Export != nil {
			// Clean-up the export.
			source.Export.Clean()
		}
		// Remove the in store source.
		delete(c.sources, hr.ResourceID().String())
		source.Unlock()
	}

	if hrs, err := c.getHelmReleasesForMirror(source.Mirror); err == nil && len(hrs) == 0 {
		// The mirror is no longer in use by any source;
		// stop and delete the mirror.
		c.mirrors.StopOne(source.Mirror)
		ok = true
	}
	return ok
}

// SyncMirrors instructs all git mirrors to sync from their respective upstreams.
func (c *GitChartSourceSync) SyncMirrors() {
	c.logger.Log("info", "starting git chart source sync")
	for _, err := range c.mirrors.RefreshAll(c.config.GitTimeout) {
		c.logger.Log("error", "failed syncing git mirror", "err", err)
	}
	c.logger.Log("info", "finished syncing git chart sources")
}

// syncGitChartSource synchronizes the source in store with the latest
// HEAD for the configured ref in the given repository. But only if it
// has seen commits for paths we are interested in. It returns the
// synchronized source or nil and a boolean indicating if the source
// was updated.
func (c *GitChartSourceSync) syncGitChartSource(repo *git.Repo, hr *v1.HelmRelease) (*GitChartSource, bool) {

	mirrorName := mirrorName(hr)
	if mirrorName == "" {
		return nil, false
	}

	source := hr.Spec.GitChartSource
	nsClient := c.client.HelmV1().HelmReleases(hr.Namespace)
	logger := log.With(c.logger,
		"mirror", source.GitURL, "ref", source.RefOrDefault(), "path", source.Path, "resource", hr.ResourceID().String())

	if repoStatus, err := repo.Status(); repoStatus != git.RepoReady {
		logger.Log("warning", "mirror not ready for sync", "status", repoStatus)
		_ = status.SetCondition(nsClient, *hr, status.NewCondition(
			v1.HelmReleaseChartFetched,
			corev1.ConditionUnknown,
			ReasonGitNotReady,
			"mirror not ready for sync: "+err.Error(),
		))
		// Repository is not ready yet; wait for signal.
		return nil, false
	}

	// Acquire sources lock and attempt to find a source in store.
	c.sourcesMu.Lock()
	s, ok := c.sources[hr.ResourceID().String()]
	if !ok {
		// No source found in store;
		// create the boilerplate of a new one.
		s = &GitChartSource{Mutex: sync.Mutex{}, Mirror: mirrorName, Remote: source.GitURL}
		c.sources[hr.ResourceID().String()] = s
	}
	// Acquire source lock and unlock sources lock.
	s.Lock()
	defer s.Unlock()
	c.sourcesMu.Unlock()

	// Get the current HEAD for the configured ref.
	ctx, cancel := context.WithTimeout(context.Background(), c.config.GitTimeout)
	refHead, err := repo.Revision(ctx, hr.Spec.GitChartSource.RefOrDefault())
	cancel()
	if err != nil {
		logger.Log("error", "failed to get current HEAD for configured ref", "err", err.Error())
		_ = status.SetCondition(nsClient, *hr, status.NewCondition(
			v1.HelmReleaseChartFetched,
			corev1.ConditionFalse,
			ReasonGitNotReady,
			"failed to get current HEAD for configured ref: "+err.Error(),
		))
		return nil, false
	}

	if ok {
		// We did find an existing source earlier; check if the
		// repository has seen commits in paths we are interested in.
		ctx, cancel = context.WithTimeout(context.Background(), c.config.GitTimeout)
		commits, err := repo.CommitsBetween(ctx, s.Head, refHead, hr.Spec.Path)
		cancel()
		// No commits or failed retrieving commit range,
		// return without updating source.
		if err != nil || len(commits) == 0 {
			if err != nil {
				_ = status.SetCondition(nsClient, *hr, status.NewCondition(
					v1.HelmReleaseChartFetched,
					corev1.ConditionFalse,
					ReasonGitNotReady,
					"failed to get commits between old and new revision: "+err.Error(),
				))
				logger.Log("error", "failed to get commits between old and new revision", "curRev", s.Head, "newRev", refHead, "err", err.Error())
			}
			return s, false
		}
	}

	// Export a new working clone for the source at the recorded ref HEAD.
	ctx, cancel = context.WithTimeout(context.Background(), c.config.GitTimeout)
	newExport, err := repo.Export(ctx, refHead)
	cancel()
	if err != nil {
		_ = status.SetCondition(nsClient, *hr, status.NewCondition(
			v1.HelmReleaseChartFetched,
			corev1.ConditionFalse,
			ReasonGitNotReady,
			"failed to clone from mirror at given revision: "+err.Error(),
		))
		logger.Log("error", "failed to clone from mirror at given revision", "rev", refHead, "err", err.Error())
		// Failed creating a git clone at the given ref HEAD.
		return s, false
	}

	if oldExport := s.Export; oldExport != nil {
		// Defer clean-up of old export.
		defer oldExport.Clean()
	}

	defer func() {
		logger.Log("info", "succesfully cloned git repository")
		status.SetCondition(nsClient, *hr, status.NewCondition(
			v1.HelmReleaseChartFetched,
			corev1.ConditionTrue,
			ReasonGitCloned,
			"successfully cloned git repository",
		))
	}()

	// :magic:
	s.Export = newExport
	s.Ref = hr.Spec.RefOrDefault()
	s.Head = refHead
	return s, true
}

// maybeMirror requests a new mirror for the given remote. The return value
// indicates whether the repo was already present (`true` if so,
// `false` otherwise).
func (c *GitChartSourceSync) maybeMirror(mirrorName string, remote string) bool {
	ok := c.mirrors.Mirror(
		mirrorName, git.Remote{URL: remote}, git.Timeout(c.config.GitTimeout),
		git.PollInterval(c.config.GitPollInterval), git.ReadOnly)
	if !ok {
		c.logger.Log("info", "started mirroring new remote", "remote", remote, "mirror", mirrorName)
	}
	return ok
}

// getHelmReleasesForMirror returns a slice of `HelmRelease`s that make
// use of the given mirror.
func (c *GitChartSourceSync) getHelmReleasesForMirror(mirror string) ([]*v1.HelmRelease, error) {
	hrs, err := c.lister.List(labels.Everything())
	if err != nil {
		return nil, err
	}
	mHrs := make([]*v1.HelmRelease, 0)
	for _, hr := range hrs {
		if mirrorName(hr) == "" {
			continue
		}
		mHrs = append(mHrs, hr.DeepCopy()) // to prevent modifying the (shared) lister store
	}
	return mHrs, nil
}

// mirrorName returns the name of the mirror for the given
// `v1.HelmRelease`.
// TODO(michael): this will not always be the git URL; e.g.
// per namespace, per auth.
func mirrorName(hr *v1.HelmRelease) string {
	if hr == nil || hr.Spec.GitChartSource == nil {
		return ""
	}
	return hr.Spec.GitURL
}
