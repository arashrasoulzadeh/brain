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
}

func (p CapacityProfile) Validate() error {
	if p.CPUCores <= 0 || p.MemoryBytes <= 0 || p.PlayersPerRoom < 1 || p.PlayersPerRoom > 8 || p.RepresentativeAppRequests <= 0 || p.UploadRequests <= 0 || p.BackgroundJobs <= 0 || p.EyeExportJobs <= 0 || p.SimulatedRTTMs < 0 {
		return ErrInvalidCapacityProfile
	}
	return nil
}
