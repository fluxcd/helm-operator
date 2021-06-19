package release

import (
	"testing"

	"github.com/stretchr/testify/assert"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"sigs.k8s.io/yaml"

	v1 "github.com/fluxcd/helm-operator/pkg/apis/helm.fluxcd.io/v1"
	"github.com/fluxcd/helm-operator/pkg/helm"
)

func TestComposeValues(t *testing.T) {
	defaultNamespace := "flux"
	otherNamespace := "other-namespace"

	client := fake.NewSimpleClientset(
		&corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "release-configmap",
				Namespace: defaultNamespace,
			},
			Data: map[string]string{
				"values.yaml": `valuesDict:
  same-namespace-configmap: true`,
			},
		},
		&corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "release-configmap",
				Namespace: otherNamespace,
			},
			Data: map[string]string{
				"values.yaml": `valuesDict:
  cross-namespace-configmap: true`,
			},
		},
		&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "release-secret",
				Namespace: defaultNamespace,
			},
			Data: map[string][]byte{
				"values.yaml": []byte(`valuesDict:
  same-namespace-secret: true`),
			},
		},
		&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "release-secret",
				Namespace: otherNamespace,
			},
			Data: map[string][]byte{
				"values.yaml": []byte(`valuesDict:
  cross-namespace-secret: true`),
			},
		},
	)

	cases := []struct {
		description      string
		releaseNamespace string
		valuesFromSource []v1.ValuesFromSource
		assertions       []func(*testing.T, helm.Values)
	}{
		{
			description:      "simple same-namespace test",
			releaseNamespace: defaultNamespace,
			valuesFromSource: []v1.ValuesFromSource{
				{
					ConfigMapKeyRef: &v1.OptionalConfigMapKeySelector{
						ConfigMapKeySelector: v1.ConfigMapKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "release-configmap",
							},
							Key: "values.yaml",
						},
					},
				},
				{
					SecretKeyRef: &v1.OptionalSecretKeySelector{
						SecretKeySelector: v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "release-secret",
							},
							Key: "values.yaml",
						},
					},
				},
			},
			assertions: []func(*testing.T, helm.Values){
				func(t *testing.T, values helm.Values) {
					assert.NotNil(t, values["valuesDict"].(map[string]interface{})["same-namespace-configmap"])
				},
				func(t *testing.T, values helm.Values) {
					assert.NotNil(t, values["valuesDict"].(map[string]interface{})["same-namespace-secret"])
				},
			},
		},
		{
			description:      "simple cross-namespace test",
			releaseNamespace: defaultNamespace,
			valuesFromSource: []v1.ValuesFromSource{
				{
					ConfigMapKeyRef: &v1.OptionalConfigMapKeySelector{
						ConfigMapKeySelector: v1.ConfigMapKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "release-configmap",
							},
							Namespace: otherNamespace,
							Key:       "values.yaml",
						},
					},
				},
				{
					SecretKeyRef: &v1.OptionalSecretKeySelector{
						SecretKeySelector: v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "release-secret",
							},
							Namespace: otherNamespace,
							Key:       "values.yaml",
						},
					},
				},
			},
			assertions: []func(*testing.T, helm.Values){
				func(t *testing.T, values helm.Values) {
					assert.NotNil(t, values["valuesDict"].(map[string]interface{})["cross-namespace-configmap"])
				},
				func(t *testing.T, values helm.Values) {
					assert.NotNil(t, values["valuesDict"].(map[string]interface{})["cross-namespace-secret"])
				},
			},
		},
		{
			description:      "same and cross-namespace test",
			releaseNamespace: defaultNamespace,
			valuesFromSource: []v1.ValuesFromSource{
				{
					ConfigMapKeyRef: &v1.OptionalConfigMapKeySelector{
						ConfigMapKeySelector: v1.ConfigMapKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "release-configmap",
							},
							Key: "values.yaml",
						},
					},
				},
				{
					SecretKeyRef: &v1.OptionalSecretKeySelector{
						SecretKeySelector: v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "release-secret",
							},
							Key: "values.yaml",
						},
					},
				},
				{
					ConfigMapKeyRef: &v1.OptionalConfigMapKeySelector{
						ConfigMapKeySelector: v1.ConfigMapKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "release-configmap",
							},
							Namespace: otherNamespace,
							Key:       "values.yaml",
						},
					},
				},
				{
					SecretKeyRef: &v1.OptionalSecretKeySelector{
						SecretKeySelector: v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "release-secret",
							},
							Namespace: otherNamespace,
							Key:       "values.yaml",
						},
					},
				},
			},
			assertions: []func(*testing.T, helm.Values){
				func(t *testing.T, values helm.Values) {
					assert.NotNil(t, values["valuesDict"].(map[string]interface{})["cross-namespace-configmap"])
				},
				func(t *testing.T, values helm.Values) {
					assert.NotNil(t, values["valuesDict"].(map[string]interface{})["cross-namespace-secret"])
				},
				func(t *testing.T, values helm.Values) {
					assert.NotNil(t, values["valuesDict"].(map[string]interface{})["cross-namespace-configmap"])
				},
				func(t *testing.T, values helm.Values) {
					assert.NotNil(t, values["valuesDict"].(map[string]interface{})["cross-namespace-secret"])
				},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			hr := &v1.HelmRelease{
				Spec: v1.HelmReleaseSpec{
					ValuesFrom: c.valuesFromSource,
				},
			}
			hr.Namespace = c.releaseNamespace

			values, err := composeValues(client.CoreV1(), hr, "", "")
			t.Log(values)
			assert.NoError(t, err)
			for _, assertion := range c.assertions {
				var hv helm.Values
				yaml.Unmarshal(values, &hv)
				assertion(t, hv)
			}
		})
	}
}
func TestComposeValuesWithAnchors(t *testing.T) {

	client := fake.NewSimpleClientset(
		&corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "anchor-configmap",
				Namespace: "flux",
			},
			Data: map[string]string{
				"values.yaml": `reference: ^&anchor success
dereference: ^^anchor`,
			},
		}, &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "nil-configmap",
				Namespace: "flux",
			},
			Data: map[string]string{
				"values.yaml": `reference:
dereference: ^^anchor`,
			},
		},
	)

	cases := []struct {
		description      string
		shouldFail       bool
		releaseNamespace string
		valuesFromSource []v1.ValuesFromSource
		assertions       []func(*testing.T, helm.Values)
		anchors          string
	}{
		{
			description:      "anchor substitution pass test",
			releaseNamespace: "flux",
			shouldFail:       false,
			valuesFromSource: []v1.ValuesFromSource{
				{
					ConfigMapKeyRef: &v1.OptionalConfigMapKeySelector{
						ConfigMapKeySelector: v1.ConfigMapKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "anchor-configmap",
							},
							Key: "values.yaml",
						},
					},
				},
			},
			anchors: "^&|^^",
			assertions: []func(*testing.T, helm.Values){
				func(t *testing.T, values helm.Values) {
					assert.Equal(t, values["dereference"], "success", "Derefrence should be success")
				},
			},
		}, {
			description:      "anchor substitution fail test",
			releaseNamespace: "flux",
			shouldFail:       true,
			valuesFromSource: []v1.ValuesFromSource{
				{
					ConfigMapKeyRef: &v1.OptionalConfigMapKeySelector{
						ConfigMapKeySelector: v1.ConfigMapKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "anchor-configmap",
							},
							Key: "values.yaml",
						},
					},
				},
			},
			anchors: "^&",
			assertions: []func(*testing.T, helm.Values){
				func(t *testing.T, values helm.Values) {

				},
			},
		}, {
			description:      "nil handling",
			releaseNamespace: "flux",
			shouldFail:       false,
			valuesFromSource: []v1.ValuesFromSource{
				{
					ConfigMapKeyRef: &v1.OptionalConfigMapKeySelector{
						ConfigMapKeySelector: v1.ConfigMapKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "nil-configmap",
							},
							Key: "values.yaml",
						},
					},
				},
			},
			anchors: "doesnt|matter",
			assertions: []func(*testing.T, helm.Values){
				func(t *testing.T, values helm.Values) {

				},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			hr := &v1.HelmRelease{
				Spec: v1.HelmReleaseSpec{
					ValuesFrom: c.valuesFromSource,
				},
			}
			hr.Namespace = c.releaseNamespace

			values, err := composeValues(client.CoreV1(), hr, "", c.anchors)
			t.Log(values)
			if c.shouldFail {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			for _, assertion := range c.assertions {
				var hv helm.Values
				yaml.Unmarshal(values, &hv)
				assertion(t, hv)
			}
		})
	}
}
