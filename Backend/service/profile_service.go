package service

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	pprofprofile "github.com/google/pprof/profile"

	"gomonitor/backend/model"
	"gomonitor/backend/pkg/profiler"
)

type ProfileService struct {
	profiler *profiler.Profiler
}

func NewProfileService(p *profiler.Profiler) *ProfileService {
	return &ProfileService{profiler: p}
}

func (s *ProfileService) CaptureSummary(req model.ProfileCaptureRequest) (model.ProfileCaptureResponse, error) {
	profileType := strings.TrimSpace(strings.ToLower(req.Type))
	if profileType == "" {
		profileType = "cpu"
	}

	result, err := s.profiler.Capture(profileType, req.Seconds)
	if err != nil {
		return model.ProfileCaptureResponse{}, err
	}

	summary, err := summarizeProfile(result.Type, result.Data)
	if err != nil {
		return model.ProfileCaptureResponse{}, err
	}

	response := model.ProfileCaptureResponse{
		ProfileType:     result.Type,
		CapturedAt:      time.Now().UTC().Format(time.RFC3339),
		DurationSeconds: result.DurationSeconds,
		SampleType:      summary.SampleType,
		SampleUnit:      summary.SampleUnit,
		TotalSamples:    summary.TotalSamples,
		TopFunctions:    summary.TopFunctions,
		DownloadURL:     fmt.Sprintf("/profiles/%s/download", result.Type),
	}

	if result.Type == "cpu" {
		response.DownloadURL = fmt.Sprintf("/profiles/%s/download?seconds=%d", result.Type, result.DurationSeconds)
	}

	if result.Type == "goroutine" {
		states, stateErr := summarizeGoroutineStates()
		if stateErr == nil {
			response.GoroutineStates = states
		}
	}

	response.Notes = buildNotes(result.Type, summary.TotalSamples)

	return response, nil
}

func (s *ProfileService) CaptureRaw(profileType string, seconds int) (profiler.CaptureResult, error) {
	return s.profiler.Capture(strings.ToLower(strings.TrimSpace(profileType)), seconds)
}

type profileSummary struct {
	SampleType   string
	SampleUnit   string
	TotalSamples int64
	TopFunctions []model.ProfileTopFunction
}

func summarizeProfile(profileType string, data []byte) (profileSummary, error) {
	parsed, err := pprofprofile.ParseData(data)
	if err != nil {
		return profileSummary{}, err
	}
	if len(parsed.SampleType) == 0 || len(parsed.Sample) == 0 {
		return profileSummary{SampleType: "samples", SampleUnit: "count"}, nil
	}

	sampleIndex := sampleIndexForType(profileType, parsed)
	topMap := make(map[string]int64)
	var total int64

	for _, sample := range parsed.Sample {
		if sampleIndex >= len(sample.Value) {
			continue
		}
		value := sample.Value[sampleIndex]
		if value < 0 {
			value = -value
		}
		total += value

		name := leafFunctionName(sample)
		if name == "" {
			name = "unknown"
		}
		topMap[name] += value
	}

	topFunctions := make([]model.ProfileTopFunction, 0, len(topMap))
	for name, value := range topMap {
		topFunctions = append(topFunctions, model.ProfileTopFunction{Name: name, Value: value})
	}
	sort.Slice(topFunctions, func(i, j int) bool {
		return topFunctions[i].Value > topFunctions[j].Value
	})
	if len(topFunctions) > 12 {
		topFunctions = topFunctions[:12]
	}

	st := parsed.SampleType[sampleIndex]
	sampleType := st.Type
	if sampleType == "" {
		sampleType = "samples"
	}
	unit := st.Unit
	if unit == "" {
		unit = "count"
	}

	return profileSummary{
		SampleType:   sampleType,
		SampleUnit:   unit,
		TotalSamples: total,
		TopFunctions: topFunctions,
	}, nil
}

func sampleIndexForType(profileType string, p *pprofprofile.Profile) int {
	if profileType != "heap" && profileType != "allocs" {
		return 0
	}

	for i, st := range p.SampleType {
		if profileType == "heap" && st.Type == "inuse_space" {
			return i
		}
		if profileType == "allocs" && st.Type == "alloc_space" {
			return i
		}
	}

	return 0
}

func leafFunctionName(sample *pprofprofile.Sample) string {
	for _, loc := range sample.Location {
		for _, line := range loc.Line {
			if line.Function != nil && line.Function.Name != "" {
				return line.Function.Name
			}
		}
	}
	return ""
}

func summarizeGoroutineStates() ([]model.ProfileStateCount, error) {
	text, err := profiler.CaptureGoroutineDebugText()
	if err != nil {
		return nil, err
	}

	stateRegex := regexp.MustCompile(`goroutine \d+ \[([^\]]+)\]:`)
	matches := stateRegex.FindAllStringSubmatch(text, -1)
	if len(matches) == 0 {
		return nil, errors.New("no goroutine states found")
	}

	counts := map[string]int{}
	for _, m := range matches {
		state := strings.TrimSpace(m[1])
		if state == "" {
			state = "unknown"
		}
		counts[state]++
	}

	states := make([]model.ProfileStateCount, 0, len(counts))
	for state, count := range counts {
		states = append(states, model.ProfileStateCount{State: state, Count: count})
	}
	sort.Slice(states, func(i, j int) bool {
		return states[i].Count > states[j].Count
	})
	return states, nil
}

func buildNotes(profileType string, totalSamples int64) []string {
	notes := make([]string, 0, 2)
	if totalSamples == 0 {
		notes = append(notes, "No samples collected. Increase load or capture duration.")
	}
	if profileType == "mutex" || profileType == "block" {
		notes = append(notes, "Mutex and block profiles depend on runtime sampling rates and workload.")
	}
	if profileType == "cpu" {
		notes = append(notes, "CPU profiling adds overhead while capture is active.")
	}
	return notes
}
