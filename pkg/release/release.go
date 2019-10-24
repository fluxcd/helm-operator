package release

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	"github.com/go-kit/kit/log"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/kubernetes"
	k8sclientv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/getter"
	k8shelm "k8s.io/helm/pkg/helm"
	helmenv "k8s.io/helm/pkg/helm/environment"
	hapi_release "k8s.io/helm/pkg/proto/hapi/release"
	helmutil "k8s.io/helm/pkg/releaseutil"

	fluxk8s "github.com/fluxcd/flux/pkg/cluster/kubernetes"
	helmfluxv1 "github.com/fluxcd/helm-operator/pkg/apis/helm.fluxcd.io/v1"
)

type Action string

const (
	InstallAction Action = "CREATE"
	UpgradeAction Action = "UPDATE"
)

// Release contains clients needed to provide functionality related to helm releases
type Release struct {
	logger     log.Logger
	HelmClient *k8shelm.Client
}

type Releaser interface {
	GetUpgradableRelease(name string) (*hapi_release.Release, error)
	Install(dir string, releaseName string, hr helmfluxv1.HelmRelease, action Action, opts InstallOptions) (*hapi_release.Release, error)
}

type DeployInfo struct {
	Name string
}

type InstallOptions struct {
	DryRun    bool
	ReuseName bool
}

// New creates a new Release instance.
func New(logger log.Logger, helmClient *k8shelm.Client) *Release {
	r := &Release{
		logger:     logger,
		HelmClient: helmClient,
	}
	return r
}

// GetUpgradableRelease returns a release if the current state of it
// allows an upgrade, a descriptive error if it is not allowed, or
// nil if the release does not exist.
func (r *Release) GetUpgradableRelease(name string) (*hapi_release.Release, error) {
	rls, err := r.HelmClient.ReleaseContent(name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, nil
		}
		return nil, err
	}

	release := rls.GetRelease()
	status := release.GetInfo().GetStatus()

	switch status.GetCode() {
	case hapi_release.Status_DEPLOYED:
		return release, nil
	case hapi_release.Status_FAILED:
		return nil, fmt.Errorf("release requires a rollback before it can be upgraded (%s)", status.GetCode().String())
	case hapi_release.Status_PENDING_INSTALL,
		hapi_release.Status_PENDING_UPGRADE,
		hapi_release.Status_PENDING_ROLLBACK:
		return nil, fmt.Errorf("operation pending for release (%s)", status.GetCode().String())
	default:
		return nil, fmt.Errorf("current state prevents it from being upgraded (%s)", status.GetCode().String())
	}
}

// shouldRollback determines if a release should be rolled back
// based on the status of the Helm release.
func (r *Release) shouldRollback(name string) (bool, error) {
	rls, err := r.HelmClient.ReleaseStatus(name)
	if err != nil {
		return false, err
	}

	status := rls.GetInfo().GetStatus()
	switch status.Code {
	case hapi_release.Status_FAILED:
		r.logger.Log("info", "rolling back release", "release", name)
		return true, nil
	case hapi_release.Status_PENDING_ROLLBACK:
		r.logger.Log("info", "release already has a rollback pending", "release", name)
		return false, nil
	default:
		return false, fmt.Errorf("release with status %s cannot be rolled back", status.Code.String())
	}
}

func (r *Release) canDelete(name string) (bool, error) {
	rls, err := r.HelmClient.ReleaseStatus(name)

	if err != nil {
		return false, err
	}
	/*
		"UNKNOWN":          0,
		"DEPLOYED":         1,
		"DELETED":          2,
		"SUPERSEDED":       3,
		"FAILED":           4,
		"DELETING":         5,
		"PENDING_INSTALL":  6,
		"PENDING_UPGRADE":  7,
		"PENDING_ROLLBACK": 8,
	*/
	status := rls.GetInfo().GetStatus()
	switch status.Code {
	case 1, 4:
		r.logger.Log("info", fmt.Sprintf("Deleting release %s", name))
		return true, nil
	case 2:
		r.logger.Log("info", fmt.Sprintf("Release %s already deleted", name))
		return false, nil
	default:
		return false, fmt.Errorf("release %s with status %s cannot be deleted", name, status.Code.String())
	}
}

// Install performs a Chart release given the directory containing the
// charts, and the HelmRelease specifying the release. Depending
// on the release type, this is either a new release, or an upgrade of
// an existing one.
//
// TODO(michael): cloneDir is only relevant if installing from git;
// either split this procedure into two varieties, or make it more
// general and calculate the path to the chart in the caller.
func (r *Release) Install(chartPath, releaseName string, hr helmfluxv1.HelmRelease, action Action, opts InstallOptions,
	kubeClient *kubernetes.Clientset) (release *hapi_release.Release, checksum string, err error) {

	defer func(start time.Time) {
		ObserveRelease(
			start,
			action,
			opts.DryRun,
			err == nil,
			hr.Namespace,
			hr.ReleaseName(),
		)
	}(time.Now())

	if chartPath == "" {
		return nil, "", fmt.Errorf("empty path to chart supplied for resource %q", hr.ResourceID().String())
	}
	_, err = os.Stat(chartPath)
	switch {
	case os.IsNotExist(err):
		return nil, "", fmt.Errorf("no file or dir at path to chart: %s", chartPath)
	case err != nil:
		return nil, "", fmt.Errorf("error statting path given for chart %s: %s", chartPath, err.Error())
	}

	r.logger.Log("info", fmt.Sprintf("processing release %s (as %s)", hr.ReleaseName(), releaseName),
		"action", fmt.Sprintf("%v", action),
		"options", fmt.Sprintf("%+v", opts),
		"timeout", fmt.Sprintf("%vs", hr.GetTimeout()))

	vals, err := Values(kubeClient.CoreV1(), hr.Namespace, chartPath, hr.GetValuesFromSources(), hr.Spec.Values)
	if err != nil {
		r.logger.Log("error", fmt.Sprintf("Failed to compose values for Chart release [%s]: %v", hr.Spec.ReleaseName, err))
		return nil, "", err
	}

	strVals, err := vals.YAML()
	if err != nil {
		r.logger.Log("error", fmt.Sprintf("Problem with supplied customizations for Chart release [%s]: %v", hr.Spec.ReleaseName, err))
		return nil, "", err
	}
	rawVals := []byte(strVals)
	checksum = ValuesChecksum(rawVals)

	switch action {
	case InstallAction:
		res, err := r.HelmClient.InstallRelease(
			chartPath,
			hr.GetTargetNamespace(),
			k8shelm.ValueOverrides(rawVals),
			k8shelm.ReleaseName(releaseName),
			k8shelm.InstallDryRun(opts.DryRun),
			k8shelm.InstallReuseName(opts.ReuseName),
			k8shelm.InstallTimeout(hr.GetTimeout()),
		)

		if err != nil {
			r.logger.Log("error", fmt.Sprintf("Chart release failed: %s: %#v", hr.Spec.ReleaseName, err))
			// purge the release if the install failed but only if this is the first revision
			history, err := r.HelmClient.ReleaseHistory(releaseName, k8shelm.WithMaxHistory(2))
			if err == nil && len(history.Releases) == 1 && history.Releases[0].Info.Status.Code == hapi_release.Status_FAILED {
				r.logger.Log("info", fmt.Sprintf("Deleting failed release: [%s]", hr.Spec.ReleaseName))
				_, err = r.HelmClient.DeleteRelease(releaseName, k8shelm.DeletePurge(true))
				if err != nil {
					r.logger.Log("error", fmt.Sprintf("Release deletion error: %#v", err))
					return nil, "", err
				}
			}
			return nil, checksum, err
		}
		if !opts.DryRun {
			r.annotateResources(res.Release, hr)
			RecordRelease()
		}
		return res.Release, checksum, err
	case UpgradeAction:
		res, err := r.HelmClient.UpdateRelease(
			releaseName,
			chartPath,
			k8shelm.UpdateValueOverrides(rawVals),
			k8shelm.UpgradeDryRun(opts.DryRun),
			k8shelm.UpgradeTimeout(hr.GetTimeout()),
			k8shelm.ResetValues(hr.Spec.ResetValues),
			k8shelm.UpgradeForce(hr.Spec.ForceUpgrade),
			k8shelm.UpgradeWait(hr.Spec.Rollback.Enable),
		)

		if err != nil {
			r.logger.Log("error", fmt.Sprintf("Chart upgrade release failed: %s: %#v", hr.Spec.ReleaseName, err))
			return nil, checksum, err
		}
		if !opts.DryRun {
			r.annotateResources(res.Release, hr)
			RecordRelease()
		}
		return res.Release, checksum, err
	default:
		err = fmt.Errorf("Valid install options: CREATE, UPDATE. Provided: %s", action)
		r.logger.Log("error", err.Error())
		return nil, "", err
	}
}

// Rollback rolls back a Chart release if required
func (r *Release) Rollback(releaseName string, hr helmfluxv1.HelmRelease) (*hapi_release.Release, error) {
	ok, err := r.shouldRollback(releaseName)
	if !ok {
		if err != nil {
			return nil, err
		}
		return nil, nil
	}

	res, err := r.HelmClient.RollbackRelease(
		releaseName,
		k8shelm.RollbackVersion(0), // '0' makes Helm fetch the latest deployed release
		k8shelm.RollbackTimeout(hr.Spec.Rollback.GetTimeout()),
		k8shelm.RollbackForce(hr.Spec.Rollback.Force),
		k8shelm.RollbackRecreate(hr.Spec.Rollback.Recreate),
		k8shelm.RollbackDisableHooks(hr.Spec.Rollback.DisableHooks),
		k8shelm.RollbackWait(hr.Spec.Rollback.Wait),
		k8shelm.RollbackDescription("Automated rollback by Helm operator"),
	)
	if err != nil {
		r.logger.Log("error", fmt.Sprintf("failed to rollback release: %#v", err))
		return nil, err
	}

	r.annotateResources(res.Release, hr)
	r.logger.Log("info", "rolled back release", "release", releaseName)

	return res.Release, err
}

// Delete purges a Chart release
func (r *Release) Delete(name string) error {
	ok, err := r.canDelete(name)
	if !ok {
		if err != nil {
			return err
		}
		return nil
	}

	_, err = r.HelmClient.DeleteRelease(name, k8shelm.DeletePurge(true))
	if err != nil {
		r.logger.Log("error", fmt.Sprintf("Release deletion error: %#v", err))
		return err
	}
	r.logger.Log("info", fmt.Sprintf("Release deleted: [%s]", name))
	return nil
}

// OwnedByHelmRelease validates the release is managed by the given
// HelmRelease, by looking for the resource ID in the antecedent
// annotation. This validation is necessary because we can not
// validate the uniqueness of a release name on the creation of a
// HelmRelease, which would result in the operator attempting to
// upgrade a release indefinitely when multiple HelmReleases with the
// same release name exist.
//
// To be able to migrate existing releases to a HelmRelease, empty
// (missing) annotations are handled as true / owned by.
func (r *Release) OwnedByHelmRelease(release *hapi_release.Release, hr helmfluxv1.HelmRelease) bool {
	objs := releaseManifestToUnstructured(release.Manifest, log.NewNopLogger())

	escapedAnnotation := strings.ReplaceAll(fluxk8s.AntecedentAnnotation, ".", `\.`)
	args := []string{"-o", "jsonpath={.metadata.annotations." + escapedAnnotation + "}", "get"}

	for ns, res := range namespacedResourceMap(objs, release.Namespace) {
		for _, r := range res {
			a := append(args, "--namespace", ns, r)

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			cmd := exec.CommandContext(ctx, "kubectl", a...)
			out, err := cmd.Output()
			if err != nil {
				continue
			}

			v := strings.TrimSpace(string(out))
			if v == "" {
				return true
			}
			return v == hr.ResourceID().String()
		}
	}

	return true
}

// annotateResources annotates each of the resources created (or updated)
// by the release so that we can spot them.
func (r *Release) annotateResources(release *hapi_release.Release, hr helmfluxv1.HelmRelease) {
	objs := releaseManifestToUnstructured(release.Manifest, r.logger)
	for namespace, res := range namespacedResourceMap(objs, release.Namespace) {
		args := []string{"annotate", "--overwrite"}
		args = append(args, "--namespace", namespace)
		args = append(args, res...)
		args = append(args, fluxk8s.AntecedentAnnotation+"="+hr.ResourceID().String())

		// The timeout is set to a high value as it may take some time
		// to annotate large umbrella charts.
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, "kubectl", args...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			r.logger.Log("output", string(output), "err", err)
		}
	}
}

// Values tries to resolve all given value file sources and merges
// them into one Values struct. It returns the merged Values.
func Values(corev1 k8sclientv1.CoreV1Interface, ns string, chartPath string, valuesFromSource []helmfluxv1.ValuesFromSource, values chartutil.Values) (chartutil.Values, error) {
	result := chartutil.Values{}

	for _, v := range valuesFromSource {
		var valueFile chartutil.Values

		switch {
		case v.ConfigMapKeyRef != nil:
			cm := v.ConfigMapKeyRef
			name := cm.Name
			key := cm.Key
			if key == "" {
				key = "values.yaml"
			}
			optional := cm.Optional != nil && *cm.Optional
			configMap, err := corev1.ConfigMaps(ns).Get(name, metav1.GetOptions{})
			if err != nil {
				if errors.IsNotFound(err) && optional {
					continue
				}
				return result, err
			}
			d, ok := configMap.Data[key]
			if !ok {
				if optional {
					continue
				}
				return result, fmt.Errorf("could not find key %v in ConfigMap %s/%s", key, ns, name)
			}
			if err := yaml.Unmarshal([]byte(d), &valueFile); err != nil {
				if optional {
					continue
				}
				return result, fmt.Errorf("unable to yaml.Unmarshal %v from %s in ConfigMap %s/%s", d, key, ns, name)
			}
		case v.SecretKeyRef != nil:
			s := v.SecretKeyRef
			name := s.Name
			key := s.Key
			if key == "" {
				key = "values.yaml"
			}
			optional := s.Optional != nil && *s.Optional
			secret, err := corev1.Secrets(ns).Get(name, metav1.GetOptions{})
			if err != nil {
				if errors.IsNotFound(err) && optional {
					continue
				}
				return result, err
			}
			d, ok := secret.Data[key]
			if !ok {
				if optional {
					continue
				}
				return result, fmt.Errorf("could not find key %s in Secret %s/%s", key, ns, name)
			}
			if err := yaml.Unmarshal(d, &valueFile); err != nil {
				return result, fmt.Errorf("unable to yaml.Unmarshal %v from %s in Secret %s/%s", d, key, ns, name)
			}
		case v.ExternalSourceRef != nil:
			es := v.ExternalSourceRef
			url := es.URL
			optional := es.Optional != nil && *es.Optional
			b, err := readURL(url)
			if err != nil {
				if optional {
					continue
				}
				return result, fmt.Errorf("unable to read value file from URL %s", url)
			}
			if err := yaml.Unmarshal(b, &valueFile); err != nil {
				if optional {
					continue
				}
				return result, fmt.Errorf("unable to yaml.Unmarshal %v from URL %s", b, url)
			}
		case v.ChartFileRef != nil:
			cf := v.ChartFileRef
			filePath := cf.Path
			optional := cf.Optional != nil && *cf.Optional
			f, err := readLocalChartFile(filepath.Join(chartPath, filePath))
			if err != nil {
				if optional {
					continue
				}
				return result, fmt.Errorf("unable to read value file from path %s", filePath)
			}
			if err := yaml.Unmarshal(f, &valueFile); err != nil {
				if optional {
					continue
				}
				return result, fmt.Errorf("unable to yaml.Unmarshal %v from URL %s", f, filePath)
			}
		}

		result = mergeValues(result, valueFile)
	}

	result = mergeValues(result, values)

	return result, nil
}

// ValuesChecksum calculates the SHA256 checksum of the given raw
// values.
func ValuesChecksum(rawValues []byte) string {
	hasher := sha256.New()
	hasher.Write(rawValues)
	return hex.EncodeToString(hasher.Sum(nil))
}

// Merges source and destination map, preferring values from the source Values
// This is slightly adapted from https://github.com/helm/helm/blob/2332b480c9cb70a0d8a85247992d6155fbe82416/cmd/helm/install.go#L359
func mergeValues(dest, src map[string]interface{}) map[string]interface{} {
	for k, v := range src {
		// If the key doesn't exist already, then just set the key to that value
		if _, exists := dest[k]; !exists {
			dest[k] = v
			continue
		}
		nextMap, ok := v.(map[string]interface{})
		// If it isn't another map, overwrite the value
		if !ok {
			dest[k] = v
			continue
		}
		// Edge case: If the key exists in the destination, but isn't a map
		destMap, isMap := dest[k].(map[string]interface{})
		// If the source map has a map for this key, prefer it
		if !isMap {
			dest[k] = v
			continue
		}
		// If we got to this point, it is a map in both, so merge them
		dest[k] = mergeValues(destMap, nextMap)
	}
	return dest
}

// readURL attempts to read a file from an url.
// This is slightly adapted from https://github.com/helm/helm/blob/2332b480c9cb70a0d8a85247992d6155fbe82416/cmd/helm/install.go#L552
func readURL(URL string) ([]byte, error) {
	var settings helmenv.EnvSettings
	flags := pflag.NewFlagSet("helm-env", pflag.ContinueOnError)
	settings.AddFlags(flags)
	settings.Init(flags)

	u, _ := url.Parse(URL)
	p := getter.All(settings)

	getterConstructor, err := p.ByScheme(u.Scheme)

	if err != nil {
		return []byte{}, err
	}

	getter, err := getterConstructor(URL, "", "", "")
	if err != nil {
		return []byte{}, err
	}
	data, err := getter.Get(URL)
	return data.Bytes(), err
}

// readLocalChartFile attempts to read a file from the chart path.
func readLocalChartFile(filePath string) ([]byte, error) {
	f, err := ioutil.ReadFile(filePath)
	if err != nil {
		return []byte{}, err
	}

	return f, nil
}

// releaseManifestToUnstructured turns a string containing YAML
// manifests into an array of Unstructured objects.
func releaseManifestToUnstructured(manifest string, logger log.Logger) []unstructured.Unstructured {
	manifests := helmutil.SplitManifests(manifest)
	var objs []unstructured.Unstructured
	for _, manifest := range manifests {
		bytes, err := yaml.YAMLToJSON([]byte(manifest))
		if err != nil {
			logger.Log("err", err)
			continue
		}

		if len(bytes) == 0 {
			continue
		}

		var u unstructured.Unstructured
		if err := u.UnmarshalJSON(bytes); err != nil {
			logger.Log("err", err)
			continue
		}

		// Helm charts may include list kinds, we are only interested in
		// the items on those lists.
		if u.IsList() {
			l, err := u.ToList()
			if err != nil {
				logger.Log("err", err)
				continue
			}
			objs = append(objs, l.Items...)
			continue
		}

		objs = append(objs, u)
	}
	return objs
}

// namespacedResourceMap iterates over the given objects and maps the
// resource identifier against the namespace from the object, if no
// namespace is present (either because the object kind has no namespace
// or it belongs to the release namespace) it gets mapped against the
// given release namespace.
func namespacedResourceMap(objs []unstructured.Unstructured, releaseNamespace string) map[string][]string {
	resources := make(map[string][]string)
	for _, obj := range objs {
		namespace := obj.GetNamespace()
		if namespace == "" {
			namespace = releaseNamespace
		}
		resource := obj.GetKind() + "/" + obj.GetName()
		resources[namespace] = append(resources[namespace], resource)
	}
	return resources
}
