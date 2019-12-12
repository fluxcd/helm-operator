package helm

import "sync"

// Clients holds multiple client (versions)
type Clients struct {
	sm sync.Map
}

func (cs *Clients) Add(version string, client Client) {
	cs.sm.Store(version, client)
}

func (cs *Clients) Load(version string) (Client, bool) {
	i, ok := cs.sm.Load(version)
	if !ok {
		return nil, false
	}
	c, ok := i.(Client)
	if !ok {
		return nil, false
	}
	return c, true
}

// Client is the generic interface for Client (v2 and v3) clients
type Client interface {
	UpgradeFromPath(chartPath string, releaseName string, values []byte, opts UpgradeOptions) (*Release, error)
	Status(releaseName string, opts StatusOptions) (*Release, error)
	History(releaseName string, opts HistoryOptions) ([]*Release, error)
	Rollback(releaseName string, opts RollbackOptions) (*Release, error)
	DependencyUpdate(chartPath string) error
	RepositoryIndex() error
	RepositoryAdd(name, url, username, password, certFile, keyFile, caFile string) error
	RepositoryRemove(name string) error
	RepositoryImport(path string) error
	Pull(ref, version, dest string) (string, error)
	PullWithRepoURL(repoURL, name, version, dest string) (string, error)
	Uninstall(releaseName string, opts UninstallOptions) error
	Version() string
}
