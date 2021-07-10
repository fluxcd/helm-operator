package operator

import (
	"fmt"
	"sync"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"

	helmfluxv1 "github.com/fluxcd/helm-operator/pkg/apis/helm.fluxcd.io/v1"
)

var (
	releaseQueueLength = prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
		Namespace: "flux",
		Subsystem: "helm_operator",
		Name:      "release_queue_length_count",
		Help:      "Count of release jobs waiting in the queue to be processed.",
	}, []string{})
	releaseCount = prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
		Namespace: "flux",
		Subsystem: "helm_operator",
		Name:      "release_count",
		Help:      "Count of releases managed by the operator.",
	}, []string{})
)

const (
	releaseLabelsName = "flux_helm_operator_release_labels"
	releaseLabelsHelp = "HelmRelease object labels"
)

type collectorKey struct {
	name            string
	targetNamespace string
}

// labelCollector implements prometheus.Collector interface to generate label metrics
type labelCollector struct {
	logger log.Logger

	sync.Mutex
	releases map[collectorKey]map[string]string
}

func newLabelCollector(logger log.Logger) *labelCollector {
	return &labelCollector{
		logger:   logger,
		releases: make(map[collectorKey]map[string]string),
	}
}

var (
	defaultLabels = []string{"release_name", "target_namespace"}
)

func (c *labelCollector) Describe(chan<- *stdprometheus.Desc) {
	// Return nothing to signal unchecked collector (since we have varying labels)
}
func (c *labelCollector) Collect(ch chan<- stdprometheus.Metric) {
	c.Lock()
	defer c.Unlock()

	for hr, labels := range c.releases {
		labelNames := mapKeys(labels)
		labelValues := mapValuesKeyOrdered(labels, labelNames)
		desc := stdprometheus.NewDesc(releaseLabelsName, releaseLabelsHelp, append(labelNames, defaultLabels...), nil)
		metric, err := stdprometheus.NewConstMetric(
			desc,
			stdprometheus.GaugeValue,
			1,
			append(labelValues, hr.name, hr.targetNamespace)...,
		)
		if err != nil {
			c.logger.Log("error", fmt.Sprintf("could not generate label metric: %s", err))
			continue
		}
		ch <- metric
	}
}
func (c *labelCollector) Add(hr helmfluxv1.HelmRelease) {
	c.Lock()
	defer c.Unlock()

	c.releases[newCollectorKey(hr)] = generateLabelMap(hr.Labels)
}
func (c *labelCollector) Remove(hr helmfluxv1.HelmRelease) {
	c.Lock()
	defer c.Unlock()

	delete(c.releases, newCollectorKey(hr))
}

func generateLabelMap(in map[string]string) map[string]string {
	out := make(map[string]string, len(in))
	for k, v := range in {
		out["label_"+k] = v
	}
	return out
}

func newCollectorKey(hr helmfluxv1.HelmRelease) collectorKey {
	return collectorKey{
		name:            hr.Name,
		targetNamespace: hr.GetTargetNamespace(),
	}
}
func mapKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
func mapValuesKeyOrdered(m map[string]string, orderedKeys []string) []string {
	vals := make([]string, 0, len(m))
	for _, k := range orderedKeys {
		vals = append(vals, m[k])
	}
	return vals
}
