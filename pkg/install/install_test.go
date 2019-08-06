package install

import (
	"testing"

	"github.com/instrumenta/kubeval/kubeval"
	"github.com/stretchr/testify/assert"
)

func testFillInTemplates(t *testing.T, params TemplateParameters) {
	manifests, err := FillInTemplates(params)
	assert.NoError(t, err)
	assert.Len(t, manifests, 3)

	config := &kubeval.Config{
		IgnoreMissingSchemas: true,
		KubernetesVersion:    "master",
	}
	for fileName, contents := range manifests {
		config.FileName = fileName
		validationResults, err := kubeval.Validate(contents, config)
		assert.NoError(t, err)
		for _, result := range validationResults {
			if len(result.Errors) > 0 {
				t.Errorf("found problems with manifest %s (Kind %s):\ncontent:\n%s\nerrors: %s",
					fileName,
					result.Kind,
					string(contents),
					result.Errors)
			}
		}
	}
}

func TestFillInTemplates(t *testing.T) {
	testFillInTemplates(t, TemplateParameters{
		Namespace: "flux",
	})

}

func TestFillInTemplatesNoNamespace(t *testing.T) {
	testFillInTemplates(t, TemplateParameters{
		Namespace: "",
	})
}
