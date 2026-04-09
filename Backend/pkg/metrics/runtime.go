package metrics

import "runtime"

func CurrentRuntimeStats() (goroutines int, memoryAlloc uint64) {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	return runtime.NumGoroutine(), mem.Alloc
}
