package types

type MediaResponse struct {
	ID        uint   `json:"id"`
	URL       string `json:"url"`
	Key       string `json:"key"`
	Kind      string `json:"kind"`
	MimeType  string `json:"mimeType"`
	SizeBytes int64  `json:"sizeBytes"`
}
