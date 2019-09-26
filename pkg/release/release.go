package release

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/weaveworks/flux/resource"
	"io/ioutil"
	"net/url"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	"github.com/go-kit/kit/log"
	"github.com/spf13/pflag"
	fluxk8s "github.com/weaveworks/flux/cluster/kubernetes"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/kubernetes"
	k8sclientv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/getter"
	helmenv "k8s.io/helm/pkg/helm/environment"
	helmutil "k8s.io/helm/pkg/releaseutil"

	helmfluxv1 "github.com/fluxcd/helm-operator/pkg/apis/helm.fluxcd.io/v1"
	"github.com/fluxcd/helm-operator/pkg/helm"
)

type Action string

const (
	InstallAction Action = "CREATE"
	UpgradeAction Action = "UPDATE"
)

// Release contains clients needed to provide functionality related to Helm releases
type Release struct {
	logger     log.Logger
	helmClient helm.Client
}

type InstallOptions struct {
	DryRun    bool
	ReuseName bool
}

// New creates a new Release instance.
func New(logger log.Logger, helmClient helm.Client) *Release {
	r := &Release{
		logger:     logger,
		helmClient: helmClient,
	}
	return r
}

// GetUpgradableRelease returns a release if the current state of it
// allows an upgrade, a descriptive error if it is not allowed, or
// nil if the release does not exist.
func (r *Release) GetUpgradableRelease(namespace string, releaseName string) (*helm.Release, error) {
	rel, err := r.helmClient.Status(releaseName, helm.StatusOptions{Namespace: namespace})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, nil
		}
		return nil, err
	}
	if rel.Info == nil {
		return nil, fmt.Errorf("release exists but can not determine if it can be upgraded due to missing metadata")
	}
	switch rel.Info.Status {
	case helm.StatusDeployed:
		return &rel, nil
	case helm.StatusFailed:
		return nil, fmt.Errorf("release requires a rollback before it can be upgraded (%s)", rel.Info.Status)
	case helm.StatusPendingInstall,
		helm.StatusPendingUpgrade,
		helm.StatusPendingRollback:
		return nil, fmt.Errorf("operation pending for release (%s)", rel.Info.Status)
	default:
		return nil, fmt.Errorf("current state prevents it from being upgraded (%s)", rel.Info.Status)
	}
}

// shouldRollback determines if a release should be rolled back
// based on the status of the Client release.
func (r *Release) shouldRollback(namespace string, releaseName string) (bool, error) {
	rel, err := r.helmClient.Status(releaseName, helm.StatusOptions{Namespace: namespace})
	if err != nil {
		return false, err
	}
	if rel.Info == nil {
		return false, fmt.Errorf("release exists but can not determine if it can be rolled back due to missing metadata")
	}
	switch rel.Info.Status {
	case helm.StatusFailed:
		return true, nil
	case helm.StatusPendingRollback:
		return false, nil
	default:
		return false, fmt.Errorf("release with status %s cannot be rolled back", rel.Info.Status)
	}
}

// canUninstall determines if a release can be rolled back based on
// the status of the Client release.
func (r *Release) canUninstall(namespace string, releaseName string) (bool, error) {
	rel, err := r.helmClient.Status(releaseName, helm.StatusOptions{Namespace: namespace})
	if err != nil {
		return false, err
	}
	if rel.Info == nil {
		return false, fmt.Errorf("release exists but can not determine if it can be uninstalled due to missing metadata")
	}
	switch rel.Info.Status {
	case helm.StatusDeployed, helm.StatusFailed:
		return true, nil
	case helm.StatusUninstalled:
		return false, nil
	default:
		return false, fmt.Errorf("release with status [%s] cannot be uninstalled", rel.Info.Status)
	}
}

// Install performs a Chart release given the directory containing the
// charts, and the HelmRelease specifying the release. Depending
// on the release type, this is either a new release, or an upgrade of
// an existing one.
func (r *Release) Install(chartPath, releaseName string, hr helmfluxv1.HelmRelease, action Action, opts InstallOptions,
	kubeClient *kubernetes.Clientset) (release *helm.Release, checksum string, err error) {

	defer func(start time.Time) {
		ObserveRelease(
			start,
			action,
			opts.DryRun,
			err == nil,
			hr.Namespace,
			hr.GetReleaseName(),
		)
	}(time.Now())

	r.logger.Log("info", fmt.Sprintf("processing release (as %s)", releaseName),
		"action", fmt.Sprintf("%v", action),
		"options", fmt.Sprintf("%+v", opts),
		"timeout", fmt.Sprintf("%vs", hr.GetTimeout()))

	vals, err := Values(kubeClient.CoreV1(), hr.Namespace, chartPath, hr.GetValuesFromSources(), hr.Spec.Values)
	if err != nil {
		r.logger.Log("error", fmt.Sprintf("failed to compose values for chart release: %v", err))
		return nil, "", err
	}

	strVals, err := vals.YAML()
	if err != nil {
		r.logger.Log("error", fmt.Sprintf("problem with supplied customizations for chart release: %v", err))
		return nil, "", err
	}
	rawVals := []byte(strVals)
	checksum = ValuesChecksum(rawVals)

	switch action {
	case InstallAction:
		rel, err := r.helmClient.InstallFromPath(chartPath, releaseName, rawVals, helm.InstallOptions{
			Namespace: hr.GetTargetNamespace(), DryRun: opts.DryRun, Replace: opts.ReuseName, Timeout: hr.GetTimeout()})
		if err != nil {
			// Delete the chart if it is an initial install;
			// this can potentially be replaced by setting the Atomic
			// install option to true, but there is a bug that causes
			// the release to also be removed if it is _not_ an initial
			// install: https://github.com/helm/helm/issues/5875
			r.logger.Log("error", fmt.Sprintf("chart release failed: %v", err))
			{
				history, err := r.helmClient.History(releaseName, helm.HistoryOptions{Namespace: hr.GetTargetNamespace(), Max: 2})
				if err == nil && len(history) == 1 && history[0].Info != nil && history[0].Info.Status == helm.StatusFailed {
					r.logger.Log("info", "cleaning up failed initial release")
					if err = r.helmClient.Uninstall(releaseName, helm.UninstallOptions{Namespace: hr.GetTargetNamespace()}); err != nil {
						r.logger.Log("error", err.Error())
						return nil, "", err
					}
				}
			}
			return nil, checksum, err
		}
		if !opts.DryRun {
			r.annotateResources(rel, hr.ResourceID())
		}
		return &rel, checksum, err
	case UpgradeAction:
		rel, err := r.helmClient.UpgradeFromPath(chartPath, releaseName, rawVals, helm.UpgradeOptions{
			Namespace:   hr.GetTargetNamespace(),
			DryRun:      opts.DryRun,
			ResetValues: hr.Spec.ResetValues,
			Force:       hr.Spec.ForceUpgrade,
			Wait:        hr.Spec.Rollback.Enable,
			Timeout:     hr.GetTimeout()})
		if err != nil {
			r.logger.Log("error", fmt.Sprintf("chart upgrade release failed: %v", err))
			return nil, checksum, err
		}
		if !opts.DryRun {
			r.annotateResources(rel, hr.ResourceID())
		}
		return &rel, checksum, err
	default:
		err = fmt.Errorf("valid install options: CREATE, UPDATE; provided: %s", action)
		r.logger.Log("error", err.Error())
		return nil, "", err
	}
}

// Rollback rolls back a Chart release if required
func (r *Release) Rollback(hr helmfluxv1.HelmRelease) (*helm.Release, error) {
	ok, err := r.shouldRollback(hr.GetTargetNamespace(), hr.GetReleaseName())
	if !ok {
		if err != nil {
			return nil, err
		}
		return nil, nil
	}
	r.logger.Log("info", "rolling back release")
	rel, err := r.helmClient.Rollback(hr.GetReleaseName(), helm.RollbackOptions{
		Namespace:    hr.GetTargetNamespace(),
		Timeout:      hr.Spec.Rollback.GetTimeout(),
		Wait:         hr.Spec.Rollback.Wait,
		DisableHooks: hr.Spec.Rollback.DisableHooks,
		Recreate:     hr.Spec.Rollback.Recreate,
		Force:        hr.Spec.Rollback.Force,
	})
	if err != nil {
		r.logger.Log("error", fmt.Sprintf("failed to rollback release: %v", err))
		return nil, err
	}

	r.annotateResources(rel, hr.ResourceID())
	r.logger.Log("info", "rolled back release")
	return &rel, err
}

// Uninstall purges a Chart release
func (r *Release) Uninstall(hr helmfluxv1.HelmRelease) error {
	ok, err := r.canUninstall(hr.GetTargetNamespace(), hr.GetReleaseName())
	if !ok {
		if err != nil {
			return err
		}
		return nil
	}

	err = r.helmClient.Uninstall(hr.GetReleaseName(), helm.UninstallOptions{Namespace: hr.GetTargetNamespace()})
	if err != nil {
		r.logger.Log("error", fmt.Sprintf("uninstall error: %v", err))
		return err
	}
	r.logger.Log("info", "uninstalled release")
	return nil
}

// ManagedByHelmRelease validates the release is managed by the given
// HelmRelease, by looking for the resource ID in the antecedent
// annotation. This validation is necessary because we can not
// validate the uniqueness of a release name on the creation of a
// HelmRelease, which would result in the operator attempting to
// upgrade a release indefinitely when multiple HelmReleases with the
// same release name exist.
//
// To be able to migrate existing releases to a HelmRelease, empty
// (missing) annotations are handled as true / owned by.
func (r *Release) ManagedByHelmRelease(release *helm.Release, hr helmfluxv1.HelmRelease) bool {
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

	return false
}

// annotateResources annotates each of the resources created (or updated)
// by the release so that we can spot them.
func (r *Release) annotateResources(rel helm.Release, resourceID resource.ID) {
	objs := releaseManifestToUnstructured(rel.Manifest, r.logger)
	for namespace, res := range namespacedResourceMap(objs, rel.Namespace) {
		args := []string{"annotate", "--overwrite"}
		args = append(args, "--namespace", namespace)
		args = append(args, res...)
		args = append(args, fluxk8s.AntecedentAnnotation+"="+resourceID.String())

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
func Values(corev1 k8sclientv1.CoreV1Interface, ns string, chartPath string,
	valuesFromSource []helmfluxv1.ValuesFromSource, values chartutil.Values) (chartutil.Values, error) {

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
	flags := pflag.NewFlagSet("helmClient-env", pflag.ContinueOnError)
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
