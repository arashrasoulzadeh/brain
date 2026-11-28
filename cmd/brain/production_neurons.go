package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"runtime"
	"strings"
	"time"

	"brain/internal/cortex"
	"brain/internal/execution"
	"brain/internal/neuron"
	"brain/internal/plugin"
	"brain/internal/synapse"
)

// startExecutionCluster selects one explicit production or managed development
// architecture. Partial production enrollment settings never select legacy.
func startExecutionCluster(httpAddress string) (*devCluster, error) {
	mode := os.Getenv("BRAIN_PLUGIN_EXECUTION")
	_, managed := os.LookupEnv("BRAIN_DEV_WORKERS")
	devMode := os.Getenv("BRAIN_DEV_PLUGIN_EXECUTION")
	configured := false
	for _, name := range []string{"BRAIN_SYNAPSE_CERT", "BRAIN_SYNAPSE_KEY", "BRAIN_SYNAPSE_NEURON_CA", "BRAIN_SYNAPSE_ENROLLMENT"} {
		if _, ok := os.LookupEnv(name); ok {
			configured = true
		}
	}
	if !managed && devMode == "" && os.Getenv("BRAIN_SYNAPSE_ADDRESS") != "" {
		configured = true
	}
	if managed || devMode != "" {
		if mode != "" || configured {
			return nil, errors.New("managed development cannot use production Synapse configuration")
		}
		return startDevCluster(nil, httpAddress)
	}
	switch mode {
	case "cortex":
		return startProductionCluster()
	case "", "legacy":
		if configured {
			return nil, errors.New("production Synapse configuration requires BRAIN_PLUGIN_EXECUTION=cortex")
		}
		return nil, nil
	default:
		return nil, errors.New("BRAIN_PLUGIN_EXECUTION must be cortex or legacy")
	}
}

func startProductionCluster() (*devCluster, error) {
	if runtime.GOOS != "linux" || os.Getenv("BRAIN_LOCAL_DEV") == "1" || os.Getenv("BRAIN_LOCAL_MACOS") == "1" {
		return nil, errors.New("production Cortex requires Linux and non-development configuration")
	}
	address := os.Getenv("BRAIN_SYNAPSE_ADDRESS")
	certPath, keyPath := os.Getenv("BRAIN_SYNAPSE_CERT"), os.Getenv("BRAIN_SYNAPSE_KEY")
	caPath, enrollmentPath := os.Getenv("BRAIN_SYNAPSE_NEURON_CA"), os.Getenv("BRAIN_SYNAPSE_ENROLLMENT")
	_, port, err := net.SplitHostPort(address)
	if err != nil || port == "" || port == "0" || certPath == "" || keyPath == "" || caPath == "" || enrollmentPath == "" {
		return nil, errors.New("production Cortex requires explicit Synapse address, certificate, key, Neuron CA and enrollment file")
	}
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, fmt.Errorf("Synapse server certificate unavailable: %w", err)
	}
	pem, err := os.ReadFile(caPath)
	if err != nil {
		return nil, fmt.Errorf("Neuron CA unavailable: %w", err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(pem) {
		return nil, errors.New("invalid Neuron CA")
	}
	raw, err := os.ReadFile(enrollmentPath)
	if err != nil {
		return nil, fmt.Errorf("Neuron enrollment unavailable: %w", err)
	}
	var enrolled map[string][]string
	if len(raw) > 1<<20 || json.Unmarshal(raw, &enrolled) != nil {
		return nil, errors.New("invalid Neuron enrollment file")
	}
	tlsConfig, err := synapse.EnrolledServerTLS(cert, pool, enrolled)
	if err != nil {
		return nil, err
	}
	runtimeDigest, err := execution.RuntimeSourceDigest(".")
	if err != nil {
		return nil, fmt.Errorf("runtime source identity unavailable: %w", err)
	}
	catalog, err := execution.NewCatalog(128, execution.MaxBundleBytes)
	if err != nil {
		return nil, err
	}
	scheduler, err := cortex.New(catalog, cortex.Config{MaxNeurons: len(enrolled), MaxProjectRunning: 32, HeartbeatTTL: 8 * time.Second})
	if err != nil {
		return nil, err
	}
	server, err := synapse.NewEnrolledGRPCServer(tlsConfig, scheduler, catalog, 2*time.Second, runtimeDigest)
	if err != nil {
		return nil, err
	}
	listener, err := net.Listen("tcp", address)
	if err != nil {
		server.Stop()
		return nil, err
	}
	cluster := &devCluster{server: server, scheduler: scheduler, catalog: catalog,
		pluginRuntime: execution.Runtime{Kind: plugin.ScheduledTypeScriptKind, Profile: neuron.LinuxProfile}}
	go func() {
		if err := server.Serve(listener); err != nil && !strings.Contains(err.Error(), "the server has been stopped") {
			log.Print("production Synapse listener stopped")
		}
	}()
	log.Printf("Production Synapse listening on %s; execution awaits enrolled Neurons", listener.Addr())
	return cluster, nil
}
