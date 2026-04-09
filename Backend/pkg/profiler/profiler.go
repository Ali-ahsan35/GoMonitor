package profiler

import (
	"bytes"
	"errors"
	"fmt"
	"runtime"
	runtimepprof "runtime/pprof"
	"sync"
	"time"
)

var errUnsupportedType = errors.New("unsupported profile type")

type CaptureResult struct {
	Type            string
	DurationSeconds int
	Data            []byte
}

type Profiler struct {
	mu sync.Mutex
}

func New() *Profiler {
	return &Profiler{}
}

func SupportedTypes() []string {
	return []string{"cpu", "heap", "allocs", "goroutine", "mutex", "block", "threadcreate"}
}

func (p *Profiler) Capture(profileType string, seconds int) (CaptureResult, error) {
	switch profileType {
	case "cpu":
		return p.captureCPU(seconds)
	case "heap", "allocs", "goroutine", "mutex", "block", "threadcreate":
		return p.captureLookup(profileType)
	default:
		return CaptureResult{}, fmt.Errorf("%w: %s", errUnsupportedType, profileType)
	}
}

func (p *Profiler) captureCPU(seconds int) (CaptureResult, error) {
	if seconds <= 0 {
		seconds = 10
	}
	if seconds > 60 {
		seconds = 60
	}

	var buf bytes.Buffer

	p.mu.Lock()
	defer p.mu.Unlock()

	if err := runtimepprof.StartCPUProfile(&buf); err != nil {
		return CaptureResult{}, err
	}
	time.Sleep(time.Duration(seconds) * time.Second)
	runtimepprof.StopCPUProfile()

	return CaptureResult{
		Type:            "cpu",
		DurationSeconds: seconds,
		Data:            buf.Bytes(),
	}, nil
}

func (p *Profiler) captureLookup(profileType string) (CaptureResult, error) {
	var buf bytes.Buffer
	if profileType == "heap" || profileType == "allocs" {
		// Trigger GC so heap profiles better represent live memory.
		runtime.GC()
	}

	profile := runtimepprof.Lookup(profileType)
	if profile == nil {
		return CaptureResult{}, fmt.Errorf("profile not available: %s", profileType)
	}
	if err := profile.WriteTo(&buf, 0); err != nil {
		return CaptureResult{}, err
	}

	return CaptureResult{
		Type:            profileType,
		DurationSeconds: 0,
		Data:            buf.Bytes(),
	}, nil
}

func CaptureGoroutineDebugText() (string, error) {
	var buf bytes.Buffer
	profile := runtimepprof.Lookup("goroutine")
	if profile == nil {
		return "", errors.New("goroutine profile not available")
	}
	if err := profile.WriteTo(&buf, 2); err != nil {
		return "", err
	}
	return buf.String(), nil
}
