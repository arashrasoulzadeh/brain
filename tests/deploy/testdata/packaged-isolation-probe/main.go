package main

import (
	"os"

	"brain/internal/plugin"
)

type packagedWorker struct{}

func (packagedWorker) Start() error { return nil }
func (packagedWorker) Stop() error  { return nil }
func (packagedWorker) Ready() bool  { return true }
func (packagedWorker) Handle(request plugin.WorkerRequest) plugin.WorkerResponse {
	return plugin.WorkerResponse{Protocol: request.Protocol, Success: true}
}

func main() {
	if len(os.Args) != 3 {
		os.Exit(2)
	}
	request := plugin.WorkerRequest{
		Protocol:  1,
		ProjectID: os.Args[2],
		PluginID:  "packaged-probe",
		Version:   "1.0.0",
		Operation: "probe",
	}
	if _, err := plugin.InvokeForProject(packagedWorker{}, request, 1, os.Args[1]); err != nil {
		os.Exit(1)
	}
}
