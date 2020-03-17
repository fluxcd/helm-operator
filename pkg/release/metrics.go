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
	LabelAction      = "action"
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
	releaseActionDuration = prometheus.NewHistogramFrom(stdprometheus.HistogramOpts{
		Namespace:   "flux",
		Subsystem:   "helm_operator",
		Name:        "release_phase_duration_seconds",
		Help:        "Release phase duration in seconds.",
		ConstLabels: nil,
		Buckets:     durationBuckets,
	}, []string{LabelAction, LabelSuccess, LabelNamespace, LabelReleaseName})
)

func ObserveRelease(start time.Time, success bool, namespace, releaseName string) {
	releaseDuration.With(
		LabelSuccess, fmt.Sprint(success),
		LabelNamespace, namespace,
		LabelReleaseName, releaseName,
	).Observe(time.Since(start).Seconds())
}

func ObserveReleaseAction(start time.Time, action action, success bool, namespace, releaseName string) {
	releaseActionDuration.With(
		LabelAction, string(action),
		LabelSuccess, fmt.Sprint(success),
		LabelNamespace, namespace,
		LabelReleaseName, releaseName,
	).Observe(time.Since(start).Seconds())
}
