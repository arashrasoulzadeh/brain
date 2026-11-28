package cortex

import (
	"slices"
)

// NeuronPerformance describes the bounded cluster-wide recent execution window,
// not lifetime counters or live process utilization. Zero resource measurements
// are unreported; sample counts distinguish missing measurements from zeroes.
type NeuronPerformance struct {
	ID string
	Completed, Succeeded, Failed, Cancelled, Lost int
	AverageWaitMillis, AverageDurationMillis, P95DurationMillis float64
	CPUSamples, MemorySamples int
	TotalCPUSeconds, AverageCPUSeconds float64
	AveragePeakMemoryBytes float64
	MaxPeakMemoryBytes int64
}

// PerformanceReport includes connected identities and identities represented in
// retained history. Its size is bounded by the snapshot inventory and history.
func (s Snapshot) PerformanceReport() []NeuronPerformance {
	type accumulator struct {
		NeuronPerformance
		wait, duration, memory float64
		durations []int64
	}
	byID := make(map[string]*accumulator)
	get := func(id string) *accumulator {
		if byID[id] == nil {
			byID[id] = &accumulator{NeuronPerformance: NeuronPerformance{ID: id}}
		}
		return byID[id]
	}
	for _, n := range s.Neurons {
		get(n.ID)
	}
	for _, row := range s.Recent {
		a := get(row.NeuronID)
		a.Completed++
		switch row.Outcome {
		case "succeeded":
			a.Succeeded++
		case "failed":
			a.Failed++
		case "cancelled":
			a.Cancelled++
		case "lost":
			a.Lost++
		}
		a.wait += float64(row.WaitMillis)
		a.duration += float64(row.DurationMillis)
		a.durations = append(a.durations, row.DurationMillis)
		if row.CPUTimeNanos > 0 {
			a.CPUSamples++
			a.TotalCPUSeconds += float64(row.CPUTimeNanos) / 1e9
		}
		if row.PeakMemoryBytes > 0 {
			a.MemorySamples++
			a.memory += float64(row.PeakMemoryBytes)
			a.MaxPeakMemoryBytes = max(a.MaxPeakMemoryBytes, row.PeakMemoryBytes)
		}
	}
	out := make([]NeuronPerformance, 0, len(byID))
	for _, a := range byID {
		if a.Completed > 0 {
			a.AverageWaitMillis = a.wait / float64(a.Completed)
			a.AverageDurationMillis = a.duration / float64(a.Completed)
			slices.Sort(a.durations)
			a.P95DurationMillis = float64(a.durations[(95*a.Completed+99)/100-1])
		}
		if a.CPUSamples > 0 {
			a.AverageCPUSeconds = a.TotalCPUSeconds / float64(a.CPUSamples)
		}
		if a.MemorySamples > 0 {
			a.AveragePeakMemoryBytes = a.memory / float64(a.MemorySamples)
		}
		out = append(out, a.NeuronPerformance)
	}
	slices.SortFunc(out, func(a, b NeuronPerformance) int {
		if a.ID < b.ID {
			return -1
		}
		if a.ID > b.ID {
			return 1
		}
		return 0
	})
	return out
}
