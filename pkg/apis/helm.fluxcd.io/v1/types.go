package v1

// +kubebuilder:validation:Enum="True";"False";"Unknown"
type ConditionStatus string

// These are valid condition statuses.
// "ConditionTrue" means a resource is in the condition,
// "ConditionFalse" means a resource is not in the condition,
// "ConditionUnknown" means the operator can't decide if a
// resource is in the condition or not.
const (
	ConditionTrue    ConditionStatus = "True"
	ConditionFalse   ConditionStatus = "False"
	ConditionUnknown ConditionStatus = "Unknown"
)

type LocalObjectReference struct {
	Name string `json:"name"`
}

type ConfigMapKeySelector struct {
	LocalObjectReference `json:",inline"`
	// +optional
	Namespace string `json:"namespace,omitempty"`
	// +optional
	Key string `json:"key,omitempty"`
}

type OptionalConfigMapKeySelector struct {
	ConfigMapKeySelector `json:",inline"`
	// +optional
	Optional bool `json:"optional,omitempty"`
}

type SecretKeySelector struct {
	LocalObjectReference `json:",inline"`
	// +optional
	Namespace string `json:"namespace,omitempty"`
	// +optional
	Key string `json:"key,omitempty"`
}

type OptionalSecretKeySelector struct {
	SecretKeySelector `json:",inline"`
	// +optional
	Optional bool `json:"optional,omitempty"`
}
