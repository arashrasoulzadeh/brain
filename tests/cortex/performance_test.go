package cortex_test

import (
	"reflect"
	"testing"

	"brain/internal/cortex"
)

func TestPerformanceReportUsesMeasuredSamplesAndNearestRank(t *testing.T) {
	snapshot := cortex.Snapshot{
		Neurons: []cortex.NeuronSnapshot{{ID: "idle"}, {ID: "active"}},
		Recent: []cortex.ExecutionTrace{
			{NeuronID: "active", Outcome: "succeeded", WaitMillis: 2, DurationMillis: 10, CPUTimeNanos: 1000000000, PeakMemoryBytes: 1024},
			{NeuronID: "active", Outcome: "failed", WaitMillis: 4, DurationMillis: 30, CPUTimeNanos: 3000000000, PeakMemoryBytes: 3072},
			{NeuronID: "active", Outcome: "cancelled", WaitMillis: 0, DurationMillis: 20},
			{NeuronID: "offline", Outcome: "lost", DurationMillis: 4},
		},
	}
	want := []cortex.NeuronPerformance{
		{ID: "active", Completed: 3, Succeeded: 1, Failed: 1, Cancelled: 1, AverageWaitMillis: 2, AverageDurationMillis: 20, P95DurationMillis: 30, CPUSamples: 2, MemorySamples: 2, TotalCPUSeconds: 4, AverageCPUSeconds: 2, AveragePeakMemoryBytes: 2048, MaxPeakMemoryBytes: 3072},
		{ID: "idle"},
		{ID: "offline", Completed: 1, Lost: 1, AverageDurationMillis: 4, P95DurationMillis: 4},
	}
	if got := snapshot.PerformanceReport(); !reflect.DeepEqual(got, want) {
		t.Fatalf("performance report = %+v, want %+v", got, want)
	}
	if got := (cortex.Snapshot{}).PerformanceReport(); len(got) != 0 {
		t.Fatal(got)
	}
}
