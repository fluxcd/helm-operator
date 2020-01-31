package v3

import (
	helm2to3v2 "github.com/helm/helm-2to3/pkg/v2"
	helm2to3v3 "github.com/helm/helm-2to3/pkg/v3"
)

type Converter struct {
	TillerNamespace string
	TillerLabel     string
}

type ConvertOptions struct {
	StorageType     string
	Max             int
	Delete          bool
}

// Convert attempts to convert the given release name from v2 to v3.
func (c Converter) Convert(releaseName string, opts ConvertOptions) error {
	retrieveOpts := helm2to3v2.RetrieveOptions{
		ReleaseName:      releaseName,
		StorageType:      opts.StorageType,
		TillerLabel:      c.TillerLabel,
		TillerNamespace:  c.TillerNamespace,
		TillerOutCluster: false,
	}
	retrieveOpts.ReleaseName = releaseName
	v2Releases, err := helm2to3v2.GetReleaseVersions(retrieveOpts)
	if err != nil {
		return err
	}

	relCount, start := len(v2Releases), 0
	if opts.Max > 0 && opts.Max < relCount {
		start = relCount - opts.Max
	}

	for i := start; i < relCount; i++ {
		v2Release := v2Releases[i]
		v3Release, err := helm2to3v3.CreateRelease(v2Release)
		if err != nil {
			return err
		}
		if err := helm2to3v3.StoreRelease(v3Release); err != nil {
			return err
		}
	}

	if opts.Delete {
		if err := helm2to3v2.DeleteAllReleaseVersions(retrieveOpts, false); err != nil {
			return err
		}
	}
	return nil
}
