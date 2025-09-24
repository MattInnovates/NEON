package structs

type Event struct {
	T int64 `json:"t"`
	Type string `json:"type"`
	Payload map[string]any `json:"payload"`
}
