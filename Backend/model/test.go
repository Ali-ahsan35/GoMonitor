package model

type RunTestRequest struct {
	URLs    []string          `json:"urls"`
	Headers map[string]string `json:"headers,omitempty"`
}

type APIResult struct {
	URL     string `json:"url"`
	TimeMS  int64  `json:"time"`
	Status  int    `json:"status"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type TestSummary struct {
	TotalTimeMS  int64  `json:"total_time"`
	SuccessCount int    `json:"success_count"`
	FailureCount int    `json:"failure_count"`
	Goroutines   int    `json:"goroutines"`
	MemoryAlloc  uint64 `json:"memory_alloc"`
}

type RunTestResponse struct {
	Results []APIResult `json:"results"`
	Summary TestSummary `json:"summary"`
}
