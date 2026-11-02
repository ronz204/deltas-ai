package diffing

type Baseline struct {
	Head  string            `json:"head"`
	Dirty map[string]string `json:"dirty"`
}
