package api

// These classes are manually created for this PoC, but it should be generated from OpenAPI specification.
type HttpResponse[T any] struct {
	Status string `json:"status"`          // "ok" | "error"
	Data   *T     `json:"data,omitempty"`  // operation‑specific payload
	Error  string `json:"error,omitempty"` // error description when Status=="error"
}
type PackageVideoRequest struct {
	FileName      string `json:"file_name"`      // e.g. "movie.mp4"
	SegmentLength int    `json:"segment_length"` // seconds per segment
}

type PackageVideoResponse struct {
	FileName string `json:"file_name,omitempty"`
	MpdURL   string `json:"mpd_url,omitempty"`
}
