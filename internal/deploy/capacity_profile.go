package deploy

import "errors"

var ErrInvalidCapacityProfile = errors.New("invalid capacity profile")

type CapacityProfile struct {
	CPUCores                  int `json:"cpu_cores"`
	MemoryBytes               int `json:"memory_bytes"`
	PlayersPerRoom            int `json:"players_per_room"`
	RepresentativeAppRequests int `json:"representative_app_requests"`
	UploadRequests            int `json:"upload_requests"`
	BackgroundJobs            int `json:"background_jobs"`
	EyeExportJobs             int `json:"eye_export_jobs"`
	SimulatedRTTMs            int `json:"simulated_rtt_ms"`
	PacketLossPercent         int `json:"packet_loss_percent"`
	P95InputSnapshotLatencyMs int `json:"p95_input_snapshot_latency_ms"`
	TickDelayMs               int `json:"tick_delay_ms"`
	MeasuredCPUPercent        int `json:"measured_cpu_percent"`
	MeasuredMemoryBytes       int `json:"measured_memory_bytes"`
	QueueDepth                int `json:"queue_depth"`
	Disconnects               int `json:"disconnects"`
	TrafficBytes              int `json:"traffic_bytes"`
}

func (p CapacityProfile) Validate() error {
	if p.CPUCores <= 0 || p.MemoryBytes <= 0 || p.PlayersPerRoom < 1 || p.PlayersPerRoom > 8 || p.RepresentativeAppRequests <= 0 || p.UploadRequests <= 0 || p.BackgroundJobs <= 0 || p.EyeExportJobs <= 0 || p.SimulatedRTTMs < 0 || p.PacketLossPercent < 0 || p.PacketLossPercent > 100 || p.P95InputSnapshotLatencyMs <= 0 || p.TickDelayMs <= 0 || p.MeasuredCPUPercent <= 0 || p.MeasuredCPUPercent > 100 || p.MeasuredMemoryBytes <= 0 || p.QueueDepth < 0 || p.Disconnects < 0 || p.TrafficBytes <= 0 {
		return ErrInvalidCapacityProfile
	}
	return nil
}
