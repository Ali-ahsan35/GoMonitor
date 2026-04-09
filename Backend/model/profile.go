package model

type ProfileCaptureRequest struct {
	Type    string `json:"type"`
	Seconds int    `json:"seconds,omitempty"`
}

type ProfileTopFunction struct {
	Name  string `json:"name"`
	Value int64  `json:"value"`
}

type ProfileStateCount struct {
	State string `json:"state"`
	Count int    `json:"count"`
}

type ProfileCaptureResponse struct {
	ProfileType     string               `json:"profile_type"`
	CapturedAt      string               `json:"captured_at"`
	DurationSeconds int                  `json:"duration_seconds"`
	SampleType      string               `json:"sample_type"`
	SampleUnit      string               `json:"sample_unit"`
	TotalSamples    int64                `json:"total_samples"`
	TopFunctions    []ProfileTopFunction `json:"top_functions"`
	GoroutineStates []ProfileStateCount  `json:"goroutine_states,omitempty"`
	DownloadURL     string               `json:"download_url"`
	Notes           []string             `json:"notes,omitempty"`
}
