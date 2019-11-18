package chartsync

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/weaveworks/flux/git"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"

	"github.com/fluxcd/helm-operator/pkg/apis/helm.fluxcd.io/v1"
	lister "github.com/fluxcd/helm-operator/pkg/client/listers/helm.fluxcd.io/v1"
)

// Various (final) errors.
var (
	ErrReleasesForMirror = errors.New("failed to get HelmRelease resources for mirror")
	ErrNoMirror          = errors.New("no existing git mirror found")
	ErrMirrorSync        = errors.New("failed syncing git mirror")
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

// GitChartSync syncs `sourceRef`s with their mirrors, and queues
// updates for `v1.HelmRelease`s the sync changes are relevant for.
type GitChartSync struct {
	logger log.Logger
	config GitConfig

	lister lister.HelmReleaseLister

	mirrors *git.Mirrors

	releaseSourcesMu   sync.RWMutex
	releaseSourcesByID map[string]sourceRef

	releaseQueue ReleaseQueue
}

// sourceRef is used for book keeping, so that we know when a
// signal we receive from a mirror is an actual update for a
// release, and if the source we hold is still the one referred
// to in the `v1.HelmRelease`.
type sourceRef struct {
	mirror string
	remote string
	ref    string
	head   string
}

// forHelmRelease returns true if the given `v1.HelmRelease`s
// `v1.GitChartSource` matches the sourceRef.
func (c sourceRef) forHelmRelease(hr *v1.HelmRelease) bool {
	if hr == nil || hr.Spec.GitChartSource == nil {
		return false
	}
	return c.mirror == mirrorName(hr) && c.remote == hr.Spec.GitURL && c.ref == hr.Spec.Ref
}

func NewGitChartSync(logger log.Logger,
	lister lister.HelmReleaseLister, cfg GitConfig, queue ReleaseQueue) *GitChartSync {

	return &GitChartSync{
		logger: logger,
		config: cfg,

		lister: lister,

		mirrors:            git.NewMirrors(),
		releaseSourcesByID: make(map[string]sourceRef),

		releaseQueue: queue,
	}
}

// Run starts the mirroring of git repositories, and processes mirror
// changes on signal, scheduling a release for a `HelmRelease` resource
// when the update is relevant to the release.
func (c *GitChartSync) Run(stopCh <-chan struct{}, errCh chan error, wg *sync.WaitGroup) {
	c.logger.Log("info", "starting sync of git chart sources")

	wg.Add(1)
	go func() {
		defer func() {
			c.mirrors.StopAllAndWait()
			wg.Done()
		}()

		for {
			select {
			case changed := <-c.mirrors.Changes():
				for mirrorName := range changed {
					repo, ok := c.mirrors.Get(mirrorName)

					hrs, err := c.helmReleasesForMirror(mirrorName)
					if err != nil {
						c.logger.Log("error", ErrReleasesForMirror.Error(), "mirror", mirrorName, "err", err)
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

						c.logger.Log("warning", ErrNoMirror.Error(), "mirror", mirrorName)
						for _, hr := range hrs {
							c.maybeMirror(mirrorName, hr.Spec.GitChartSource.GitURL)
						}
						// Wait for the signal from the newly requested mirror...
						continue
					}

					c.processChangedMirror(mirrorName, repo, hrs)
				}
			case <-stopCh:
				c.logger.Log("info", "stopping sync of git chart sources")
				return
			}
		}
	}()
}

// GetMirrorCopy returns a newly exported copy of the git mirror at the
// recorded HEAD and a string with the HEAD commit hash, or an error.
func (c *GitChartSync) GetMirrorCopy(hr *v1.HelmRelease) (*git.Export, string, error) {
	mirror := mirrorName(hr)
	repo, ok := c.mirrors.Get(mirror)
	if !ok {
		// We did not find a mirror; request one, return, and wait for
		// signal.
		c.maybeMirror(mirror, hr.Spec.GitURL)
		return nil, "", ChartNotReadyError{ErrNoMirror}
	}

	s, ok, err := c.sync(hr, mirror, repo)
	if err != nil {
		return nil, "", err
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.config.GitTimeout)
	defer cancel()
	export, err := repo.Export(ctx, s.head)
	if err != nil {
		return nil, "", ChartUnavailableError{err}
	}

	return export, s.head, nil
}

// Delete cleans up the source reference for the given `v1.HelmRelease`,
// this includes the mirror if there is no reference to it from sources.
// It returns a boolean indicating a successful removal (`true` if so,
// `false` otherwise).
func (c *GitChartSync) Delete(hr *v1.HelmRelease) bool {
	c.releaseSourcesMu.Lock()
	defer c.releaseSourcesMu.Unlock()

	// Attempt to get the source from store.
	source, ok := c.releaseSourcesByID[hr.ResourceID().String()]
	if ok {
		// Remove the in store source.
		delete(c.releaseSourcesByID, hr.ResourceID().String())

		if hrs, err := c.helmReleasesForMirror(source.mirror); err == nil && len(hrs) == 0 {
			// The mirror is no longer in use by any source;
			// stop and delete the mirror.
			c.mirrors.StopOne(source.mirror)
		}
	}
	return ok
}

// SyncMirrors instructs all git mirrors to sync from their respective
// upstreams.
func (c *GitChartSync) SyncMirrors() {
	c.logger.Log("info", "starting sync of git mirrors")
	for _, err := range c.mirrors.RefreshAll(c.config.GitTimeout) {
		c.logger.Log("error", ErrMirrorSync.Error(), "err", err)
	}
	c.logger.Log("info", "finished syncing git mirror")
}

// processChangedMirror syncs all given `v1.HelmRelease`s with the
// mirror we received a change signal for and schedules a release,
// but only if the sync indicated the change was relevant.
func (c *GitChartSync) processChangedMirror(mirror string, repo *git.Repo, hrs []*v1.HelmRelease) {
	for _, hr := range hrs {
		if _, ok, _ := c.sync(hr, mirror, repo); ok {
			cacheKey, err := cache.MetaNamespaceKeyFunc(hr.GetObjectMeta())
			if err != nil {
				continue // this should never happen
			}
			// Schedule release sync by adding it to the queue.
			c.releaseQueue.AddRateLimited(cacheKey)
		}
	}
}

func (c *GitChartSync) get(hr *v1.HelmRelease) (sourceRef, bool) {
	c.releaseSourcesMu.RLock()
	defer c.releaseSourcesMu.RUnlock()
	if s, ok := c.releaseSourcesByID[hr.ResourceID().String()]; ok && s.forHelmRelease(hr) {
		return s, ok
	}
	return sourceRef{}, false
}

func (c *GitChartSync) store(hr *v1.HelmRelease, s sourceRef) {
	c.releaseSourcesMu.Lock()
	c.releaseSourcesByID[hr.ResourceID().String()] = s
	c.releaseSourcesMu.Unlock()
}

// sync synchronizes the record we have for the given `v1.HelmRelease`
// with the given mirror. It always updates the HEAD record in the
// `sourceRef`, but only returns `true` if the update was relevant for
// the release (e.g. a change in git the chart source path, or a new
// record). In case of failure it returns an error.
func (c *GitChartSync) sync(hr *v1.HelmRelease, mirrorName string, repo *git.Repo) (sourceRef, bool, error) {
	source := hr.Spec.GitChartSource
	if source == nil {
		return sourceRef{}, false, nil
	}

	if status, err := repo.Status(); status != git.RepoReady {
		return sourceRef{}, false, ChartNotReadyError{err}
	}

	var changed bool
	s, ok := c.get(hr)
	if !ok {
		s = sourceRef{mirror: mirrorName, remote: source.GitURL, ref: source.RefOrDefault()}
		changed = true
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.config.GitTimeout)
	head, err := repo.Revision(ctx, s.ref)
	cancel()
	if err != nil {
		return sourceRef{}, false, ChartUnavailableError{err}
	}

	if !changed {
		// If the head still equals to what is in our books, there are no changes.
		if s.head == head {
			return s, false, nil
		}

		// Check if the mirror has seen commits in paths we are interested in for
		// this release.
		ctx, cancel = context.WithTimeout(context.Background(), c.config.GitTimeout)
		commits, err := repo.CommitsBetween(ctx, s.head, head, source.Path)
		cancel()
		if err != nil {
			return sourceRef{}, false, ChartUnavailableError{err}
		}
		changed = len(commits) > 0
	}

	s.head = head
	c.store(hr, s)
	return s, changed, nil
}

// maybeMirror requests a new mirror for the given remote. The return value
// indicates whether the repo was already present (`true` if so,
// `false` otherwise).
func (c *GitChartSync) maybeMirror(mirrorName string, remote string) bool {
	ok := c.mirrors.Mirror(
		mirrorName, git.Remote{URL: remote}, git.Timeout(c.config.GitTimeout),
		git.PollInterval(c.config.GitPollInterval), git.ReadOnly)
	if !ok {
		c.logger.Log("info", "started mirroring new remote", "remote", remote, "mirror", mirrorName)
	}
	return ok
}

// helmReleasesForMirror returns a slice of `HelmRelease`s that make
// use of the given mirror.
func (c *GitChartSync) helmReleasesForMirror(mirror string) ([]*v1.HelmRelease, error) {
	hrs, err := c.lister.List(labels.Everything())
	if err != nil {
		return nil, err
	}
	mHrs := make([]*v1.HelmRelease, 0)
	for _, hr := range hrs {
		if m := mirrorName(hr); m == "" || m != mirror {
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
	if hr != nil && hr.Spec.GitChartSource != nil {
		return hr.Spec.GitURL
	}
	return ""
}
