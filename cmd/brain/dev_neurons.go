package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"brain/internal/app"
	"brain/internal/cortex"
	"brain/internal/execution"
	"brain/internal/neuron"
	"brain/internal/plugin"
	"brain/internal/synapse"

	"google.golang.org/grpc"
)

const devBrainName = "brain.local"

type devCA struct {
	cert *x509.Certificate
	key  *ecdsa.PrivateKey
	pem  []byte
}

func newDevCA() (*devCA, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	cert := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "Brain ephemeral development CA"}, NotBefore: now.Add(-time.Minute), NotAfter: now.Add(24 * time.Hour), IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature}
	raw, err := x509.CreateCertificate(rand.Reader, cert, cert, &key.PublicKey, key)
	if err != nil {
		return nil, err
	}
	cert, err = x509.ParseCertificate(raw)
	if err != nil {
		return nil, err
	}
	return &devCA{cert: cert, key: key, pem: pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: raw})}, nil
}

func (ca *devCA) issue(id string, serial int64, server bool) ([]byte, []byte, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	now := time.Now()
	cert := &x509.Certificate{SerialNumber: big.NewInt(serial), NotBefore: now.Add(-time.Minute), NotAfter: now.Add(24 * time.Hour), KeyUsage: x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}}
	if server {
		cert.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}
		cert.DNSNames = []string{devBrainName}
	} else {
		cert.URIs = []*url.URL{{Scheme: "spiffe", Host: "brain", Path: "/neurons/" + id}}
	}
	raw, err := x509.CreateCertificate(rand.Reader, cert, ca.cert, &key.PublicKey, ca.key)
	if err != nil {
		return nil, nil, err
	}
	private, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return nil, nil, err
	}
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: raw}), pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: private}), nil
}

type devNeuronProcess struct {
	id   string
	cmd  *exec.Cmd
	done chan struct{}
	err  error
}

type devCluster struct {
	server        *grpc.Server
	scheduler     *cortex.Scheduler
	catalog       *execution.Catalog
	pluginRuntime execution.Runtime
	dir           string
	processes     []*devNeuronProcess
}

func startDevCluster(server *app.Server, httpAddress string) (*devCluster, error) {
	rawCount, enabled := os.LookupEnv("BRAIN_DEV_WORKERS")
	mode := os.Getenv("BRAIN_DEV_PLUGIN_EXECUTION")
	if mode != "" && ((mode != "cortex" && mode != "legacy") || !enabled) {
		return nil, errors.New("BRAIN_DEV_PLUGIN_EXECUTION must be cortex or legacy and requires managed local Neurons")
	}
	if !enabled {
		return nil, nil
	}
	if mode == "" {
		mode = "cortex"
	}
	count, err := strconv.Atoi(rawCount)
	if err != nil || count < 0 || count > 32 || !strings.HasPrefix(httpAddress, "127.0.0.1:") || os.Getenv("BRAIN_LOCAL_DEV") != "1" {
		return nil, errors.New("managed Neurons require loopback local development and 0–32 workers")
	}
	slots := 1
	if raw := os.Getenv("BRAIN_DEV_SLOTS"); raw != "" {
		slots, err = strconv.Atoi(raw)
		if err != nil || slots < 1 || slots > 8 {
			return nil, errors.New("BRAIN_DEV_SLOTS must be between 1 and 8")
		}
	}
	dir, err := os.MkdirTemp("", "brain-dev-synapse-")
	if err != nil {
		return nil, err
	}
	cluster := &devCluster{dir: dir}
	success := false
	defer func() {
		if !success {
			cluster.Close()
		}
	}()
	ca, err := newDevCA()
	if err != nil {
		return nil, err
	}
	caPath := filepath.Join(dir, "ca.pem")
	if err = os.WriteFile(caPath, ca.pem, 0o600); err != nil {
		return nil, err
	}
	certPEM, keyPEM, err := ca.issue("", 2, true)
	if err != nil {
		return nil, err
	}
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}
	pool := x509.NewCertPool()
	pool.AddCert(ca.cert)
	tlsConfig, err := synapse.ServerTLS(cert, pool)
	if err != nil {
		return nil, err
	}
	catalog, err := execution.NewCatalog(128, execution.MaxBundleBytes)
	if err != nil {
		return nil, err
	}
	scheduler, err := cortex.New(catalog, cortex.Config{MaxNeurons: 64, MaxProjectRunning: 32, HeartbeatTTL: 8 * time.Second})
	if err != nil {
		return nil, err
	}
	cluster.scheduler = scheduler
	cluster.catalog = catalog
	if mode == "cortex" {
		profile := neuron.LinuxProfile
		if runtime.GOOS == "darwin" {
			profile = neuron.MacOSProfile
		}
		cluster.pluginRuntime = execution.Runtime{Kind: plugin.ScheduledTypeScriptKind, Profile: profile}
	}
	cluster.server, err = synapse.NewGRPCServer(tlsConfig, scheduler, catalog, 2*time.Second)
	if err != nil {
		return nil, err
	}
	address := os.Getenv("BRAIN_SYNAPSE_ADDRESS")
	if address == "" {
		address = "127.0.0.1:0"
	}
	host, _, err := net.SplitHostPort(address)
	if err != nil || host != "127.0.0.1" {
		return nil, errors.New("development Synapse must listen on 127.0.0.1")
	}
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}
	go func() {
		if err := cluster.server.Serve(listener); err != nil {
			log.Print("development Synapse listener stopped")
		}
	}()
	binary, err := os.Executable()
	if err != nil {
		return nil, err
	}
	for i := 0; i < count; i++ {
		id := fmt.Sprintf("local-neuron-%02d", i+1)
		certPEM, keyPEM, err = ca.issue(id, int64(i)+3, false)
		if err != nil {
			return nil, err
		}
		certPath, keyPath := filepath.Join(dir, id+".pem"), filepath.Join(dir, id+".key")
		if err = os.WriteFile(certPath, certPEM, 0o600); err != nil {
			return nil, err
		}
		if err = os.WriteFile(keyPath, keyPEM, 0o600); err != nil {
			return nil, err
		}
		args := []string{"neuron", "connect", "--brain", listener.Addr().String(), "--id", id, "--cert", certPath, "--key", keyPath, "--ca", caPath, "--server-name", devBrainName, "--slots", strconv.Itoa(slots)}
		if runtime.GOOS == "darwin" {
			args = append(args, "--local-macos")
		}
		if mode == "cortex" {
			args = append(args, "--plugin-runtime")
		}
		cmd := exec.Command(binary, args...)
		cmd.Env = []string{}
		// Do not pass operator credentials, application storage configuration,
		// or other Brain process environment to a Neuron.
		// GOCOVERDIR belongs to the trusted instrumented Go process, not its
		// sandbox. The existing OS launcher clears the sandbox environment.
		for _, name := range []string{"PATH", "HOME", "TMPDIR", "GOCOVERDIR"} {
			if value, ok := os.LookupEnv(name); ok {
				cmd.Env = append(cmd.Env, name+"="+value)
			}
		}
		cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
		if err = cmd.Start(); err != nil {
			return nil, err
		}
		process := &devNeuronProcess{id: id, cmd: cmd, done: make(chan struct{})}
		cluster.processes = append(cluster.processes, process)
		go func() {
			process.err = process.cmd.Wait()
			close(process.done)
		}()
	}
	// A process is not counted until its isolated runtime probe has passed
	// and its authenticated registration is visible in Cortex.
	deadline := time.Now().Add(45 * time.Second)
	for len(scheduler.Snapshot().Neurons) != count {
		for _, process := range cluster.processes {
			select {
			case <-process.done:
				return nil, fmt.Errorf("%s exited before registration: %v", process.id, process.err)
			default:
			}
		}
		if time.Now().After(deadline) {
			return nil, errors.New("timed out waiting for development Neuron registration")
		}
		time.Sleep(100 * time.Millisecond)
	}
	if server != nil {
		server.SetNeuronScheduler(scheduler)
	}
	log.Printf("Synapse development pool: %d connected Neurons · %d slots each · %s", count, slots, listener.Addr())
	success = true
	return cluster, nil
}

func (c *devCluster) Close() {
	if c == nil {
		return
	}
	for _, process := range c.processes {
		_ = process.cmd.Process.Signal(os.Interrupt)
	}
	timer := time.NewTimer(3 * time.Second)
	defer timer.Stop()
	timedOut := false
	for _, process := range c.processes {
		if !timedOut {
			select {
			case <-process.done:
				continue
			case <-timer.C:
				timedOut = true
			}
		}
		_ = process.cmd.Process.Kill()
		<-process.done
	}
	if c.server != nil {
		c.server.Stop()
	}
	_ = os.RemoveAll(c.dir)
}

// Run a cancellation-aware heartbeat expiry sweep even without incoming work.
func (c *devCluster) monitor(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.scheduler.Expire()
		}
	}
}
