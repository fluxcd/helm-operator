package release

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"sigs.k8s.io/yaml"

	v1 "github.com/fluxcd/helm-operator/pkg/apis/helm.fluxcd.io/v1"
	"github.com/fluxcd/helm-operator/pkg/helm"
)

// composeValues attempts to compose the final values for the given
// `HelmRelease`. It returns the values as bytes and a checksum,
// or an error in case anything went wrong.
func composeValues(coreV1Client corev1client.CoreV1Interface, hr *v1.HelmRelease, chartPath string, anchorPattern string) ([]byte, error) {
	result := helm.Values{}

	for _, v := range hr.GetValuesFromSources() {
		var valueFile helm.Values
		ns := hr.Namespace

		switch {
		case v.ConfigMapKeyRef != nil:
			cm := v.ConfigMapKeyRef
			name := cm.Name
			if cm.Namespace != "" {
				ns = cm.Namespace
			}
			key := cm.Key
			if key == "" {
				key = "values.yaml"
			}
			configMap, err := coreV1Client.ConfigMaps(ns).Get(name, metav1.GetOptions{})
			if err != nil {
				if errors.IsNotFound(err) && cm.Optional {
					continue
				}
				return nil, err
			}
			d, ok := configMap.Data[key]
			if !ok {
				if cm.Optional {
					continue
				}
				return nil, fmt.Errorf("could not find key %v in ConfigMap %s/%s", key, ns, name)
			}
			if err := yaml.Unmarshal([]byte(d), &valueFile); err != nil {
				if cm.Optional {
					continue
				}
				return nil, fmt.Errorf("unable to yaml.Unmarshal %v from %s in ConfigMap %s/%s", d, key, ns, name)
			}
		case v.SecretKeyRef != nil:
			s := v.SecretKeyRef
			name := s.Name
			if s.Namespace != "" {
				ns = s.Namespace
			}
			key := s.Key
			if key == "" {
				key = "values.yaml"
			}
			secret, err := coreV1Client.Secrets(ns).Get(name, metav1.GetOptions{})
			if err != nil {
				if errors.IsNotFound(err) && s.Optional {
					continue
				}
				return nil, err
			}
			d, ok := secret.Data[key]
			if !ok {
				if s.Optional {
					continue
				}
				return nil, fmt.Errorf("could not find key %s in Secret %s/%s", key, ns, name)
			}
			if err := yaml.Unmarshal(d, &valueFile); err != nil {
				return nil, fmt.Errorf("unable to yaml.Unmarshal %v from %s in Secret %s/%s", d, key, ns, name)
			}
		case v.ExternalSourceRef != nil:
			es := v.ExternalSourceRef
			u := es.URL
			optional := es.Optional != nil && *es.Optional
			b, err := readURL(u)
			if err != nil {
				if optional {
					continue
				}
				return nil, fmt.Errorf("unable to read value file from URL %s", u)
			}
			if err := yaml.Unmarshal(b, &valueFile); err != nil {
				if optional {
					continue
				}
				return nil, fmt.Errorf("unable to yaml.Unmarshal %v from URL %s", b, u)
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
				return nil, fmt.Errorf("unable to read value file from path %s", filePath)
			}
			if err := yaml.Unmarshal(f, &valueFile); err != nil {
				if optional {
					continue
				}
				return nil, fmt.Errorf("unable to yaml.Unmarshal %v from path %s", f, filePath)
			}
		}
		result = mergeValues(result, valueFile)
	}

	result = mergeValues(result, hr.Spec.Values.Data)

	if anchorPattern != "" {
		var anchors []string
		if anchors = strings.SplitN(anchorPattern, "|", 2); len(anchors) != 2 {
			return nil, fmt.Errorf("anchor patterns is not 2, check your string and syntax docs")
		}
		references := make(map[string]string)
		dereferences := make(map[string]*string)

		final := buildAnchorMaps(result, anchors, references, dereferences)
		for k, v := range dereferences {
			if rval, ok := references[k]; ok {
				*v = rval
			}
		}
		return yaml.Marshal(final)
	}
	return result.YAML()
}

// Create an anchor if a substitution character is defined
func buildAnchorMaps(valuesYaml interface{}, anchors []string, references map[string]string, dereferences map[string]*string) interface{} {

	refrenceRegex := regexp.MustCompile(`(?:^\s*` + regexp.QuoteMeta(anchors[0]) + `)(?P<key>\S*)(?:\s*)(?P<value>.*)`)
	derefrenceRegex := regexp.MustCompile(`(?:^\s*` + regexp.QuoteMeta(anchors[1]) + `)(?P<key>\S*)`)

	newValues := make(map[string]interface{})

	if valuesYaml == nil {
		return nil
	}

	r := reflect.ValueOf(valuesYaml)
	switch r.Kind() {
	case reflect.Map:
		i := r.MapRange()
		for i.Next() {
			k := i.Key().Interface().(string)
			v := i.Value().Interface()
			newValues[k] = buildAnchorMaps(v, anchors, references, dereferences)
		}
	case reflect.String:
		s := r.Interface().(string)
		refs := refrenceRegex.FindStringSubmatch(s)
		drefs := derefrenceRegex.FindStringSubmatch(s)
		if refs != nil {
			result := make(map[string]string)
			for i, name := range refrenceRegex.SubexpNames() {
				if i != 0 && name != "" {
					result[name] = refs[i]
				}
			}
			references[result["key"]] = result["value"]
		}
		if drefs != nil {
			result := make(map[string]string)
			for i, name := range derefrenceRegex.SubexpNames() {
				if i != 0 && name != "" {
					result[name] = drefs[i]
				}
			}
			dereferences[result["key"]] = &s
			return &s
		}
		return s
	default:
		return valuesYaml
	}
	return newValues
}

// readURL attempts to read a file from an HTTP(S) URL.
func readURL(URL string) ([]byte, error) {
	u, err := url.Parse(URL)
	if err != nil {
		return []byte{}, err
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return []byte{}, fmt.Errorf("URL scheme should be HTTP(S), got '%s'", u.Scheme)
	}
	resp, err := http.Get(u.String())
	if err != nil {
		return []byte{}, err
	}
	switch resp.StatusCode {
	case http.StatusOK:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return []byte{}, err
		}
		return body, nil
	default:
		return []byte{}, fmt.Errorf("failed to retrieve file from URL, status '%s (%d)'", resp.Status, resp.StatusCode)
	}
}

// readLocalChartFile attempts to read a file from the chart path.
func readLocalChartFile(filePath string) ([]byte, error) {
	f, err := ioutil.ReadFile(filePath)
	if err != nil {
		return []byte{}, err
	}
	return f, nil
}

// mergeValues merges source and destination map, preferring values
// from the source values. This is slightly adapted from:
// https://github.com/helm/helm/blob/2332b480c9cb70a0d8a85247992d6155fbe82416/cmd/helm/install.go#L359
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
