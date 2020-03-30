package release

import (
	"fmt"
	"time"

	"github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

const (
	LabelSuccess     = "success"
	LabelNamespace   = "namespace"
	LabelReleaseName = "release_name"
)

var (
	durationBuckets = []float64{1, 5, 10, 30, 60, 120, 180, 300}
	releaseDuration = prometheus.NewHistogramFrom(stdprometheus.HistogramOpts{
		Namespace: "flux",
		Subsystem: "helm_operator",
		Name:      "release_duration_seconds",
		Help:      "Release duration in seconds.",
		Buckets:   durationBuckets,
	}, []string{LabelSuccess, LabelNamespace, LabelReleaseName})
	lastReleaseStatus = prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
		Namespace: "flux",
		Subsystem: "helm_operator",
		Name:      "last_release_status",
		Help:      "1 if last release was successful, 0 otherwise.",
	}, []string{LabelNamespace, LabelReleaseName})
)

func ObserveRelease(start time.Time, success bool, namespace, releaseName string) {
	releaseDuration.With(
		LabelSuccess, fmt.Sprint(success),
		LabelNamespace, namespace,
		LabelReleaseName, releaseName,
	).Observe(time.Since(start).Seconds())

	var success64 float64
	if success {
		success64 = 1
	}
	lastReleaseStatus.With(
		LabelNamespace, namespace,
		LabelReleaseName, releaseName,
	).Set(success64)
}
