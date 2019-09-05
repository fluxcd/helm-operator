package helm

type HistoryOptions struct {
	Namespace string `json:"namespace,omitempty"`
	Max       int    `json:"max,omitempty"`
}
