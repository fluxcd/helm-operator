package release

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	helmfluxv1 "github.com/fluxcd/helm-operator/pkg/apis/helm.fluxcd.io/v1"
)

func TestComposeValues(t *testing.T) {
	namespace := "flux"
	falseVal := false

	client := fake.NewSimpleClientset(
		&corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "release-configmap",
				Namespace: namespace,
			},
			Data: map[string]string{
				"values.yaml": `valuesDict:
  configmap: true`,
			},
		},
		&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "release-secret",
				Namespace: namespace,
			},
			Data: map[string][]byte{
				"values.yaml": []byte(`valuesDict:
  secret: true`),
			},
		},
	)

	valuesFromSource := []helmfluxv1.ValuesFromSource{
		{
			ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "release-configmap",
				},
				Key:      "values.yaml",
				Optional: &falseVal,
			},
			SecretKeyRef:      nil,
			ExternalSourceRef: nil,
			ChartFileRef:      nil,
		},
		{
			ConfigMapKeyRef: nil,
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "release-secret",
				},
				Key:      "values.yaml",
				Optional: &falseVal,
			},
			ExternalSourceRef: nil,
			ChartFileRef:      nil,
		}}

	hr := &helmfluxv1.HelmRelease{
		Spec: helmfluxv1.HelmReleaseSpec{
			ValuesFrom: valuesFromSource,
		},
	}
	hr.Namespace = namespace

	values, err := composeValues(client.CoreV1(), hr, "")
	assert.NoError(t, err)
	assert.NotNil(t, values["valuesDict"].(map[string]interface{})["configmap"])
	assert.NotNil(t, values["valuesDict"].(map[string]interface{})["secret"])
}
