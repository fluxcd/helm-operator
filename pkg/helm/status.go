package helm

type StatusOptions struct {
	Namespace string `json:"namespace,omitempty"`
	Version   int    `json:"version,omitempty"`
}
