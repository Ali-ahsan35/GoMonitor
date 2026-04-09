package service

import (
	"context"
	"sync"
	"time"

	"gomonitor/backend/model"
	"gomonitor/backend/pkg/httpclient"
	"gomonitor/backend/pkg/metrics"
)

type TestService struct {
	client *httpclient.Client
}

func NewTestService(client *httpclient.Client) *TestService {
	return &TestService{client: client}
}

func (s *TestService) RunTest(ctx context.Context, urls []string, headers map[string]string) model.RunTestResponse {
	start := time.Now()
	resultsCh := make(chan model.APIResult, len(urls))

	var wg sync.WaitGroup
	for _, rawURL := range urls {
		url := rawURL
		wg.Add(1)
		go func() {
			defer wg.Done()

			requestStart := time.Now()
			status, err := s.client.Get(ctx, url, headers)
			elapsed := time.Since(requestStart).Milliseconds()

			result := model.APIResult{
				URL:     url,
				TimeMS:  elapsed,
				Status:  status,
				Success: err == nil && status >= 200 && status < 400,
			}
			if err != nil {
				result.Error = err.Error()
			}
			resultsCh <- result
		}()
	}

	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	results := make([]model.APIResult, 0, len(urls))
	successCount := 0
	failureCount := 0

	for result := range resultsCh {
		results = append(results, result)
		if result.Success {
			successCount++
		} else {
			failureCount++
		}
	}

	goroutines, memoryAlloc := metrics.CurrentRuntimeStats()

	return model.RunTestResponse{
		Results: results,
		Summary: model.TestSummary{
			TotalTimeMS:  time.Since(start).Milliseconds(),
			SuccessCount: successCount,
			FailureCount: failureCount,
			Goroutines:   goroutines,
			MemoryAlloc:  memoryAlloc,
		},
	}
}
