package v1

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/fluxcd/flux/pkg/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AntecedentAnnotation is an annotation on a resource indicating that
// the cause of that resource is a HelmRelease. We use this rather than
// the `OwnerReference` type built into Kubernetes as this does not
// allow cross-namespace references by design. The value is expected to
// be a serialised `resource.ID`.
const AntecedentAnnotation = "helm.fluxcd.io/antecedent"

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// HelmRelease is a type to represent a Helm release.
// +k8s:openapi-gen=true
// +kubebuilder:printcolumn:name="Release",type="string",JSONPath=".status.releaseName",description="Release is the name of the Helm release, as given by Helm."
// +kubebuilder:printcolumn:name="Phase",type="string",JSONPath=".status.phase",description="Phase is the current release phase being performed for the HelmRelease."
// +kubebuilder:printcolumn:name="ReleaseStatus",type="string",JSONPath=".status.releaseStatus",description="ReleaseStatus is the status of the Helm release, as given by Helm."
// +kubebuilder:printcolumn:name="Message",type="string",JSONPath=".status.conditions[?(@.type==\"Released\")].message",description=""
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="CreationTimestamp is a timestamp representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC."
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=helmreleases,shortName=hr;hrs
type HelmRelease struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec   HelmReleaseSpec   `json:"spec"`
	Status HelmReleaseStatus `json:"status,omitempty"`
}

// ResourceID returns an ID made from the identifying parts of the
// resource, as a convenience for Flux, which uses them
// everywhere.
func (hr HelmRelease) ResourceID() resource.ID {
	return resource.MakeID(hr.Namespace, "HelmRelease", hr.Name)
}

// GetReleaseName returns the configured release name, or constructs and
// returns one based on the namespace and name of the HelmRelease.
// When the HelmRelease's metadata.namespace and spec.targetNamespace
// differ, both are used in the generated name.
// This name is used for naming and operating on the release in Helm.
func (hr HelmRelease) GetReleaseName() string {
	if hr.Spec.ReleaseName == "" {
		namespace := hr.GetDefaultedNamespace()
		targetNamespace := hr.GetTargetNamespace()

		if namespace != targetNamespace {
			// prefix the releaseName with the administering HelmRelease namespace as well
			return fmt.Sprintf("%s-%s-%s", namespace, targetNamespace, hr.Name)
		}
		return fmt.Sprintf("%s-%s", targetNamespace, hr.Name)
	}

	return hr.Spec.ReleaseName
}

// GetDefaultedNamespace returns the HelmRelease's namespace
// defaulting to the "default" if not set.
func (hr HelmRelease) GetDefaultedNamespace() string {
	if hr.GetNamespace() == "" {
		return "default"
	}
	return hr.Namespace
}

// GetTargetNamespace returns the configured release targetNamespace
// defaulting to the namespace of the HelmRelease if not set.
func (hr HelmRelease) GetTargetNamespace() string {
	if hr.Spec.TargetNamespace == "" {
		return hr.GetDefaultedNamespace()
	}
	return hr.Spec.TargetNamespace
}

func (hr HelmRelease) GetHelmVersion(defaultVersion string) string {
	if hr.Spec.HelmVersion != "" {
		return string(hr.Spec.HelmVersion)
	}
	if defaultVersion != "" {
		return defaultVersion
	}
	return string(HelmV2)
}

// GetTimeout returns the install or upgrade timeout (defaults to 300s)
func (hr HelmRelease) GetTimeout() time.Duration {
	if hr.Spec.Timeout == nil {
		return 300 * time.Second
	}
	return time.Duration(*hr.Spec.Timeout) * time.Second
}

// GetMaxHistory returns the maximum number of release
// revisions to keep (defaults to 10)
func (hr HelmRelease) GetMaxHistory() int {
	if hr.Spec.MaxHistory == nil {
		return 10
	}
	return *hr.Spec.MaxHistory
}

// GetReuseValues returns if the values of the previous release should
// be reused based on the value of `ResetValues`. When this value is
// not explicitly set, it is assumed values should not be reused, as
// this aligns with the declarative behaviour of the operator.
func (hr HelmRelease) GetReuseValues() bool {
	switch hr.Spec.ResetValues {
	case nil:
		return false
	default:
		return !*hr.Spec.ResetValues
	}
}

// GetValuesFromSources maintains backwards compatibility with
// ValueFileSecrets by merging them into the ValuesFrom array.
func (hr HelmRelease) GetValuesFromSources() []ValuesFromSource {
	valuesFrom := hr.Spec.ValuesFrom
	// Maintain backwards compatibility with ValueFileSecrets.
	if hr.Spec.ValueFileSecrets != nil {
		var secretKeyRefs []ValuesFromSource
		for _, ref := range hr.Spec.ValueFileSecrets {
			s := &OptionalSecretKeySelector{}
			s.Name = ref.Name
			secretKeyRefs = append(secretKeyRefs, ValuesFromSource{SecretKeyRef: s})
		}
		valuesFrom = append(secretKeyRefs, valuesFrom...)
	}
	return valuesFrom
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// HelmReleaseList is a list of HelmReleases
type HelmReleaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []HelmRelease `json:"items"`
}

type ChartSource struct {
	// +optional
	*GitChartSource `json:",inline"`
	// +optional
	*RepoChartSource `json:",inline"`
}

// GitChartSource describes a Helm chart sourced from Git.
type GitChartSource struct {
	// Git URL is the URL of the Git repository, e.g.
	// `git@github.com:org/repo`, `http://github.com/org/repo`,
	// or `ssh://git@example.com:2222/org/repo.git`.
	// +kubebuilder:validation:Optional
	GitURL string `json:"git"`
	// Ref is the Git branch (or other reference) to use. Defaults to
	// 'master', or the configured default Git ref.
	// +kubebuilder:validation:Optional
	Ref string `json:"ref"`
	// Path is the path to the chart relative to the repository root.
	// +kubebuilder:validation:Optional
	Path string `json:"path"`
	// SecretRef holds the authentication secret for accessing the Git
	// repository (over HTTPS). The credentials will be added to an
	// HTTPS GitURL before the mirror is started.
	// +optional
	SecretRef *LocalObjectReference `json:"secretRef,omitempty"`
	// SkipDepUpdate will tell the operator to skip running
	// 'helm dep update' before installing or upgrading the chart, the
	// chart dependencies _must_ be present for this to succeed.
	// +optional
	SkipDepUpdate bool `json:"skipDepUpdate,omitempty"`
}

// RefOrDefault returns the configured ref of the chart source. If the chart source
// does not specify a ref, the provided default is used instead.
func (s GitChartSource) RefOrDefault(defaultGitRef string) string {
	if s.Ref == "" {
		return defaultGitRef
	}
	return s.Ref
}

// RepoChartSources describes a Helm chart sourced from a Helm
// repository.
type RepoChartSource struct {
	// RepoURL is the URL of the Helm repository, e.g.
	// `https://kubernetes-charts.storage.googleapis.com` or
	// `https://charts.example.com`.
	// +kubebuilder:validation:Optional
	RepoURL string `json:"repository"`
	// Name is the name of the Helm chart _without_ an alias, e.g.
	// redis (for `helm upgrade [flags] stable/redis`).
	// +kubebuilder:validation:Optional
	Name string `json:"name"`
	// Version is the targeted Helm chart version, e.g. 7.0.1.
	// +kubebuilder:validation:Optional
	Version string `json:"version"`
	// ChartPullSecret holds the reference to the authentication secret for accessing
	// the Helm repository using HTTPS basic auth.
	// NOT IMPLEMENTED!
	// +kubebuilder:validation:Optional
	// +optional
	ChartPullSecret *LocalObjectReference `json:"chartPullSecret,omitempty"`
}

// CleanRepoURL returns the RepoURL but ensures it ends with a trailing
// slash.
func (s RepoChartSource) CleanRepoURL() string {
	cleanURL := strings.TrimRight(s.RepoURL, "/")
	return cleanURL + "/"
}

type ValuesFromSource struct {
	// The reference to a config map with release values.
	// +optional
	ConfigMapKeyRef *OptionalConfigMapKeySelector `json:"configMapKeyRef,omitempty"`
	// The reference to a secret with release values.
	// +optional
	SecretKeyRef *OptionalSecretKeySelector `json:"secretKeyRef,omitempty"`
	// The reference to an external source with release values.
	// +optional
	ExternalSourceRef *ExternalSourceSelector `json:"externalSourceRef,omitempty"`
	// The reference to a local chart file with release values.
	// +optional
	ChartFileRef *ChartFileSelector `json:"chartFileRef,omitempty"`
}

type ChartFileSelector struct {
	// Path is the file path to the source relative to the chart root.
	Path string `json:"path"`
	// Optional will mark this ChartFileSelector as optional.
	// The result of this are that operations are permitted without
	// the source, due to it e.g. being temporarily unavailable.
	// +optional
	Optional *bool `json:"optional,omitempty"`
}

type ExternalSourceSelector struct {
	// URL is the URL of the external source.
	URL string `json:"url"`
	// Optional will mark this ExternalSourceSelector as optional.
	// The result of this are that operations are permitted without
	// the source, due to it e.g. being temporarily unavailable.
	// +optional
	Optional *bool `json:"optional,omitempty"`
}

type Rollback struct {
	// Enable will mark this Helm release for rollbacks.
	// +optional
	Enable bool `json:"enable,omitempty"`
	// Retry will mark this Helm release for upgrade retries after a
	// rollback.
	// +optional
	Retry bool `json:"retry,omitempty"`
	// MaxRetries is the maximum amount of upgrade retries the operator
	// should make before bailing.
	// +optional
	MaxRetries *int64 `json:"maxRetries,omitempty"`
	// Force will mark this Helm release to `--force` rollbacks. This
	// forces the resource updates through delete/recreate if needed.
	// +optional
	Force bool `json:"force,omitempty"`
	// Recreate will mark this Helm release to `--recreate-pods` for
	// if applicable. This performs pod restarts.
	// +optional
	Recreate bool `json:"recreate,omitempty"`
	// DisableHooks will mark this Helm release to prevent hooks from
	// running during the rollback.
	// +optional
	DisableHooks bool `json:"disableHooks,omitempty"`
	// Timeout is the time to wait for any individual Kubernetes
	// operation (like Jobs for hooks) during rollback.
	// +optional
	Timeout *int64 `json:"timeout,omitempty"`
	// Wait will mark this Helm release to wait until all Pods,
	// PVCs, Services, and minimum number of Pods of a Deployment,
	// StatefulSet, or ReplicaSet are in a ready state before marking
	// the release as successful.
	// +optional
	Wait bool `json:"wait,omitempty"`
}

// GetTimeout returns the configured timout for the Helm release,
// or the default of 300s.
func (r Rollback) GetTimeout() time.Duration {
	if r.Timeout == nil {
		return 300 * time.Second
	}
	return time.Duration(*r.Timeout) * time.Second
}

// GetMaxRetries returns the configured max retries for the Helm
// release, or the default of 5.
func (r Rollback) GetMaxRetries() int64 {
	if r.MaxRetries == nil {
		return 5
	}
	return *r.MaxRetries
}

// HelmVersion is the version of Helm to target. If not supplied,
// the lowest _enabled Helm version_ will be targeted.
// Valid HelmVersion values are:
// "v2",
// "v3"
// +kubebuilder:validation:Enum="v2";"v3"
// +optional
type HelmVersion string

const (
	HelmV2 HelmVersion = "v2"
	HelmV3 HelmVersion = "v3"
)

type HelmValues struct {
	// Data holds the configuration keys and values.
	// Work around for https://github.com/kubernetes-sigs/kubebuilder/issues/528
	Data map[string]interface{} `json:"-"`
}

// MarshalJSON marshals the HelmValues data to a JSON blob.
func (v HelmValues) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.Data)
}

// UnmarshalJSON sets the HelmValues to a copy of data.
func (v *HelmValues) UnmarshalJSON(data []byte) error {
	var out map[string]interface{}
	err := json.Unmarshal(data, &out)
	if err != nil {
		return err
	}
	v.Data = out
	return nil
}

// DeepCopyInto is an deepcopy function, copying the receiver, writing
// into out. In must be non-nil. Declaring this here prevents it from
// being generated in `zz_generated.deepcopy.go`.
//
// This exists here to work around https://github.com/kubernetes/code-generator/issues/50,
// and partially around https://github.com/kubernetes-sigs/controller-tools/pull/126
// and https://github.com/kubernetes-sigs/controller-tools/issues/294.
func (in *HelmValues) DeepCopyInto(out *HelmValues) {
	b, err := json.Marshal(in.Data)
	if err != nil {
		// The marshal should have been performed cleanly as otherwise
		// the resource would not have been created by the API server.
		panic(err)
	}
	var c map[string]interface{}
	err = json.Unmarshal(b, &c)
	if err != nil {
		panic(err)
	}
	out.Data = c
	return
}

type HelmReleaseSpec struct {
	HelmVersion `json:"helmVersion,omitempty"`
	// +kubebuilder:validation:Required
	ChartSource `json:"chart"`
	// ReleaseName is the name of the The Helm release. If not supplied,
	// it will be generated by affixing the namespace to the resource
	// name.
	ReleaseName string `json:"releaseName,omitempty"`
	// MaxHistory is the maximum amount of revisions to keep for the
	// Helm release. If not supplied, it defaults to 10.
	MaxHistory *int `json:"maxHistory,omitempty"`
	// ValueFileSecrets holds the local name references to secrets.
	// DEPRECATED, use ValuesFrom.secretKeyRef instead.
	ValueFileSecrets []LocalObjectReference `json:"valueFileSecrets,omitempty"`
	ValuesFrom       []ValuesFromSource     `json:"valuesFrom,omitempty"`
	// TargetNamespace overrides the targeted namespace for the Helm
	// release. The default namespace equals to the namespace of the
	// HelmRelease resource.
	// +optional
	TargetNamespace string `json:"targetNamespace,omitempty"`
	// Timeout is the time to wait for any individual Kubernetes
	// operation (like Jobs for hooks) during installation and
	// upgrade operations.
	// +optional
	Timeout *int64 `json:"timeout,omitempty"`
	// ResetValues will mark this Helm release to reset the values
	// to the defaults of the targeted chart before performing
	// an upgrade. Not explicitly setting this to `false` equals
	// to `true` due to the declarative nature of the operator.
	// +optional
	ResetValues *bool `json:"resetValues,omitempty"`
	// SkipCRDs will mark this Helm release to skip the creation
	// of CRDs during a Helm 3 installation.
	// +optional
	SkipCRDs bool `json:"skipCRDs,omitempty"`
	// Wait will mark this Helm release to wait until all Pods,
	// PVCs, Services, and minimum number of Pods of a Deployment,
	// StatefulSet, or ReplicaSet are in a ready state before marking
	// the release as successful.
	// +optional
	Wait bool `json:"wait,omitempty"`
	// Force will mark this Helm release to `--force` upgrades. This
	// forces the resource updates through delete/recreate if needed.
	// +optional
	ForceUpgrade bool `json:"forceUpgrade,omitempty"`
	// The rollback settings for this Helm release.
	// +optional
	Rollback Rollback `json:"rollback,omitempty"`
	// Values holds the values for this Helm release.
	// +optional
	Values HelmValues `json:"values,omitempty"`
}

// HelmReleaseConditionType represents an HelmRelease condition value.
// Valid HelmReleaseConditionType values are:
// "ChartFetched",
// "Released",
// "RolledBack"
// +kubebuilder:validation:Enum="ChartFetched";"Released";"RolledBack"
// +optional
type HelmReleaseConditionType string

const (
	// ChartFetched means the chart to which the HelmRelease refers
	// has been fetched successfully.
	HelmReleaseChartFetched HelmReleaseConditionType = "ChartFetched"
	// Released means the chart release, as specified in this
	// HelmRelease, has been processed by Helm.
	HelmReleaseReleased HelmReleaseConditionType = "Released"
	// RolledBack means the chart to which the HelmRelease refers
	// has been rolled back.
	HelmReleaseRolledBack HelmReleaseConditionType = "RolledBack"
)

type HelmReleaseCondition struct {
	// Type of the condition, one of ('ChartFetched', 'Released', 'RolledBack').
	Type HelmReleaseConditionType `json:"type"`

	// Status of the condition, one of ('True', 'False', 'Unknown').
	Status ConditionStatus `json:"status"`

	// LastUpdateTime is the timestamp corresponding to the last status
	// update of this condition.
	// +optional
	LastUpdateTime *metav1.Time `json:"lastUpdateTime,omitempty"`

	// LastTransitionTime is the timestamp corresponding to the last status
	// change of this condition.
	// +optional
	LastTransitionTime *metav1.Time `json:"lastTransitionTime,omitempty"`

	// Reason is a brief machine readable explanation for the condition's last
	// transition.
	// +optional
	Reason string `json:"reason,omitempty"`

	// Message is a human readable description of the details of the last
	// transition, complementing reason.
	// +optional
	Message string `json:"message,omitempty"`
}

// HelmReleasePhase represents the phase a HelmRelease is in.
// Valid HelmReleasePhase values are:
// "ChartFetched",
// "ChartFetchFailed",
// "Installing",
// "Upgrading",
// "Succeeded",
// "Failed",
// "RollingBack",
// "RolledBack",
// "RollbackFailed",
// +kubebuilder:validation:Enum="ChartFetched";"ChartFetchFailed";"Installing";"Upgrading";"Succeeded";"Failed";"RollingBack";"RolledBack";"RollbackFailed"
// +optional
type HelmReleasePhase string

const (
	// ChartFetched means the chart to which the HelmRelease refers
	// has been fetched successfully
	HelmReleasePhaseChartFetched HelmReleasePhase = "ChartFetched"
	// ChartFetchedFailed means the chart to which the HelmRelease
	// refers could not be fetched.
	HelmReleasePhaseChartFetchFailed HelmReleasePhase = "ChartFetchFailed"

	// Installing means the installation for the HelmRelease is running.
	HelmReleasePhaseInstalling HelmReleasePhase = "Installing"
	// Upgrading means the upgrade for the HelmRelease is running.
	HelmReleasePhaseUpgrading HelmReleasePhase = "Upgrading"
	// Succeeded means the dry-run, installation, or upgrade for the
	// HelmRelease succeeded.
	HelmReleasePhaseSucceeded HelmReleasePhase = "Succeeded"
	// Failed means the installation or upgrade for the HelmRelease
	// failed.
	HelmReleasePhaseFailed HelmReleasePhase = "Failed"

	// RollingBack means a rollback for the HelmRelease is running.
	HelmReleasePhaseRollingBack HelmReleasePhase = "RollingBack"
	// RolledBack means the HelmRelease has been rolled back.
	HelmReleasePhaseRolledBack HelmReleasePhase = "RolledBack"
	// RolledBackFailed means the rollback for the HelmRelease failed.
	HelmReleasePhaseRollbackFailed HelmReleasePhase = "RollbackFailed"
)

// HelmReleaseStatus contains status information about an HelmRelease.
type HelmReleaseStatus struct {
	// ObservedGeneration is the most recent generation observed by
	// the operator.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Phase the release is in, one of ('ChartFetched',
	// 'ChartFetchFailed', 'Installing', 'Upgrading', 'Succeeded',
	// 'RollingBack', 'RolledBack', 'RollbackFailed')
	// +optional
	Phase HelmReleasePhase `json:"phase,omitempty"`

	// ReleaseName is the name as either supplied or generated.
	// +optional
	ReleaseName string `json:"releaseName,omitempty"`

	// ReleaseStatus is the status as given by Helm for the release
	// managed by this resource.
	// +optional
	ReleaseStatus string `json:"releaseStatus,omitempty"`

	// Revision holds the Git hash or version of the chart currently
	// deployed.
	// +optional
	Revision string `json:"revision,omitempty"`

	// LastAttemptedRevision is the revision of the latest chart
	// sync, and may be of a failed release.
	// +optional
	LastAttemptedRevision string `json:"lastAttemptedRevision,omitempty"`

	// RollbackCount records the amount of rollback attempts made,
	// it is incremented after a rollback failure and reset after a
	// successful upgrade or revision change.
	// +optional
	RollbackCount int64 `json:"rollbackCount,omitempty"`

	// Conditions contains observations of the resource's state, e.g.,
	// has the chart which it refers to been fetched.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []HelmReleaseCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}
