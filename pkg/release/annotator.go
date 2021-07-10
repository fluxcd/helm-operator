package release

import (
	"context"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"

	"github.com/fluxcd/flux/pkg/resource"
	"helm.sh/helm/v3/pkg/releaseutil"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"

	v1 "github.com/fluxcd/helm-operator/pkg/apis/helm.fluxcd.io/v1"
	"github.com/fluxcd/helm-operator/pkg/helm"
)

// managedByHelmRelease determines if the given `helm.Release` is
// managed by the given `v1.HelmRelease`. A release is managed when
// the resources contain a antecedent annotation with the resource ID
// of the `v1.HelmRelease`. In case the annotation is not found, we
// assume the release has been installed manually and we want to
// take over.
func managedByHelmRelease(release *helm.Release, hr v1.HelmRelease, c client.Client) (bool, string, error) {
	objs := releaseManifestToUnstructured(release.Manifest)

	errs := errCollection{}
	for _, o := range objs{
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		cancel()
		err := c.Get(ctx, client.ObjectKey{Namespace: o.GetNamespace(),Name: o.GetName()}, &o)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		id, ok := o.GetAnnotations()[v1.AntecedentAnnotation]
		if !ok{
			return false, hr.ResourceID().String(), nil
		}
		if id != hr.ResourceID().String(){
			return false, id, nil
		}
	}
	if !errs.Empty() {
		return false, "", errs
	}
	return true, hr.ResourceID().String(), nil
}

// annotateResources annotates each of the resources created (or updated)
// by the release so that we can spot them.
func annotateResources(rel *helm.Release, c client.Client, resourceID resource.ID) error {
	objs := releaseManifestToUnstructured(rel.Manifest)
	errs := errCollection{}
	for _, o := range objs {
		// merge type: application/merge-patch+json
		o.SetAnnotations(map[string]string{v1.AntecedentAnnotation: resourceID.String()})
		err := c.Patch(context.TODO(), &o, client.Merge, &client.PatchOptions{FieldManager: "apply"})
		if err != nil {
			errs = append(errs, err)
		}
	}
	if !errs.Empty() {
		return errs
	}
	return nil
}

// releaseManifestToUnstructured turns a string containing YAML
// manifests into an array of Unstructured objects.
func releaseManifestToUnstructured(manifest string) []unstructured.Unstructured {
	manifests := releaseutil.SplitManifests(manifest)
	var objs []unstructured.Unstructured
	for _, manifest := range manifests {
		var u unstructured.Unstructured

		if err := yaml.Unmarshal([]byte(manifest), &u); err != nil {
			continue
		}

		// Helm charts may include list kinds, we are only interested in
		// the items on those lists.
		if u.IsList() {
			l, err := u.ToList()
			if err != nil {
				continue
			}

			for _, i := range l.Items {
				if i.GetNamespace() == "" {
					i.SetNamespace("default")
				}
				objs = append(objs, i)
			}
		}
		if u.GetNamespace() == "" {
			u.SetNamespace("default")
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
		res := obj.GetKind() + "/" + obj.GetName()
		resources[namespace] = append(resources[namespace], res)
	}
	return resources
}
