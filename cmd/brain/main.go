package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"brain/internal/admin"
	"brain/internal/app"
	"brain/internal/config"
	"brain/internal/deploy"
	"brain/internal/operator"
	"brain/internal/plugin"
	"brain/internal/project"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "neuron" {
		if len(os.Args) < 3 || os.Args[2] != "connect" {
			log.Fatal("usage: brain neuron connect --brain host:port --id ID --cert FILE --key FILE --ca FILE --server-name NAME")
		}
		if err := neuronConnect(os.Args[3:]); err != nil {
			log.Fatal(err)
		}
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "plugin-contract" {
		if len(os.Args) != 4 {
			log.Fatal("usage: brain plugin-contract <plugin.yaml> <markdown|typescript>")
		}
		manifest, err := plugin.LoadManifest(os.Args[2])
		if err != nil {
			log.Fatal(err)
		}
		switch os.Args[3] {
		case "markdown":
			fmt.Print(plugin.Documentation(manifest))
		case "typescript":
			fmt.Print(plugin.TypeScriptContract(manifest))
		default:
			log.Fatal("unknown contract format")
		}
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "bootstrap-operator" {
		bootstrapOperator()
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "bootstrap-dashboard" {
		bootstrapDashboard()
		return
	}
	if len(os.Args) > 2 && os.Args[1] == "verify-object-manifest" {
		verifyObjectManifest(os.Args[2])
		return
	}
	cfg := config.FromEnv()
	if err := cfg.SafeError(); err != nil {
		log.Fatal(err)
	}
	if len(os.Args) > 1 && os.Args[1] == "health-only" {
		log.Print("health-only mode: no application routes; readiness is unavailable")
		if err := serveHTTP(cfg.HTTPAddress, app.NewServer().Handler()); err != nil {
			log.Fatal(err)
		}
		return
	}
	// Build the selected pool and catalog before restoring installations.
	// An unavailable configured Cortex never falls back to local workers.
	cluster, err := startExecutionCluster(cfg.HTTPAddress)
	if err != nil {
		log.Fatal(err)
	}
	server, err := productionServerWithCluster(cluster)
	if err != nil {
		cluster.Close()
		log.Fatal(err)
	}
	if cluster != nil {
		server.SetNeuronScheduler(cluster.scheduler)
	}
	if cluster != nil {
		defer cluster.Close()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go cluster.monitor(ctx)
	}
	if err := serveHTTP(cfg.HTTPAddress, server.Handler()); err != nil {
		if cluster != nil {
			cluster.Close()
		}
		log.Fatal(err)
	}
}

func productionServer() (*app.Server, error) {
	return productionServerWithCluster(nil)
}

func productionServerWithCluster(cluster *devCluster) (*app.Server, error) {
	username, secret := os.Getenv("OPERATOR_USERNAME"), os.Getenv("OPERATOR_SECRET")
	if username == "" || secret == "" {
		return nil, fmt.Errorf("OPERATOR_USERNAME and OPERATOR_SECRET are required")
	}
	if os.Getenv("RUNTIME_DSN") != "" || os.Getenv("MIGRATION_DSN") != "" {
		return nil, fmt.Errorf("PostgreSQL application storage is not wired; refusing implicit SQLite fallback")
	}
	projectsPath := os.Getenv("PROJECTS_FILE")
	if strings.TrimSpace(projectsPath) == "" || projectsPath == ":memory:" || strings.HasPrefix(projectsPath, "file:") {
		return nil, fmt.Errorf("PROJECTS_FILE must name an explicit durable SQLite file")
	}
	operators := &operator.Store{}
	digest := func(value string) string {
		sum := sha256.Sum256([]byte(value))
		return hex.EncodeToString(sum[:])
	}
	if _, err := operators.Bootstrap(username, secret, digest); err != nil {
		return nil, fmt.Errorf("operator bootstrap failed: %w", err)
	}
	registry := plugin.NewRegistry(1)
	locations, err := registerDiscoveredPlugins(registry)
	if err != nil {
		return nil, err
	}
	scheduledPlugins := cluster != nil && cluster.pluginRuntime.Kind != ""
	if !scheduledPlugins {
		if err := bindSandboxedTypeScriptWorkers(registry, locations); err != nil {
			return nil, err
		}
	}
	if err := os.MkdirAll(filepath.Dir(projectsPath), 0o700); err != nil {
		return nil, fmt.Errorf("project storage directory unavailable: %w", err)
	}
	projects, err := project.NewSQLiteStore(projectsPath)
	if err != nil {
		return nil, fmt.Errorf("project storage unavailable: %w", err)
	}
	// This database now includes credential hashes. Restrict its file mode
	// before opening the plugin storage connection.
	if err := os.Chmod(projectsPath, 0o600); err != nil {
		_ = projects.Close()
		return nil, fmt.Errorf("project storage permissions unavailable: %w", err)
	}
	storage, err := plugin.OpenSQLiteStorage(projectsPath)
	if err != nil {
		_ = projects.Close()
		return nil, fmt.Errorf("plugin storage unavailable: %w", err)
	}
	if scheduledPlugins {
		if err := storage.BindArtifactPins(cluster.catalog); err != nil {
			_ = storage.Close()
			_ = projects.Close()
			return nil, fmt.Errorf("artifact pin storage unavailable: %w", err)
		}
		if err := bindCortexPluginWorkers(registry, locations, cluster); err != nil {
			_ = storage.Close()
			_ = projects.Close()
			return nil, err
		}
	}
	if err := registry.BindSQLiteStorage(storage); err != nil {
		_ = storage.Close()
		_ = projects.Close()
		return nil, fmt.Errorf("registry state unavailable: %w", err)
	}
	registry.BindProjectResolver(func(id string) (plugin.ProjectSnapshot, error) {
		model, err := projects.Resolve(id)
		return plugin.ProjectSnapshot{ID: model.ID, Active: model.Active}, err
	})
	for _, installation := range registry.InstallationStatuses("") {
		if installation.Enabled && installation.Ready && !installation.Draining {
			if err := registry.StartWorker(installation.ProjectID, installation.PluginID, installation.Version); err != nil {
				log.Printf("restored worker unavailable: %s@%s", installation.PluginID, installation.Version)
			}
		}
	}
	go registry.RunObservationDelivery(context.Background(), time.Second)
	server := app.NewServerWithOperatorAdmin(admin.NewLifecycleServiceWithProjects(operators, registry, projects), operators, digest)
	server.SetReadinessCheck(func(ctx context.Context) error {
		if err := registry.LifecycleStorageHealth(); err != nil {
			return err
		}
		if err := projects.CheckHealth(ctx); err != nil {
			return err
		}
		if err := storage.CheckHealth(ctx); err != nil {
			return err
		}
		if cluster != nil && cluster.pluginRuntime.Kind != "" && !cluster.scheduler.HasRuntime(cluster.pluginRuntime) {
			return plugin.ErrWorkerNotReady
		}
		for _, installation := range registry.InstallationStatuses("") {
			if installation.Enabled && !installation.Draining && (!installation.Ready || !installation.WorkerReady) {
				return plugin.ErrWorkerNotReady
			}
		}
		return nil
	})
	return server, nil
}

func registerDiscoveredPlugins(registry *plugin.Registry) (map[string]plugin.PluginLocation, error) {
	registered := map[string]plugin.PluginLocation{}
	locations, err := plugin.Discover("plugins")
	if err != nil {
		return nil, fmt.Errorf("plugin discovery unavailable: %w", err)
	}
	if len(locations) == 0 {
		return nil, fmt.Errorf("no plugins discovered")
	}
	pending := make(map[string]plugin.PluginLocation, len(locations))
	manifests := make(map[string]plugin.Manifest, len(locations))
	for _, location := range locations {
		manifest, err := plugin.LoadManifest(location.ManifestPath)
		if err != nil {
			return nil, fmt.Errorf("plugin %s rejected: %w", location.ID, err)
		}
		if err := plugin.ValidatePluginSources(location.Directory, manifest); err != nil {
			return nil, fmt.Errorf("plugin %s sources rejected: %w", location.ID, err)
		}
		key := manifest.ID + "@" + manifest.Version
		if _, exists := pending[key]; exists {
			return nil, fmt.Errorf("duplicate plugin %s", key)
		}
		pending[key] = location
		manifests[key] = manifest
	}
	for len(pending) > 0 {
		progress := false
		for key, location := range pending {
			if err := registry.Register(manifests[key]); err != nil {
				if err == plugin.ErrDependencyMissing {
					continue
				}
				return nil, fmt.Errorf("plugin %s rejected: %w", key, err)
			}
			registered[key] = location
			delete(pending, key)
			progress = true
		}
		if !progress {
			return nil, fmt.Errorf("plugin discovery: %w", plugin.ErrDependencyMissing)
		}
	}
	return registered, nil
}

type sandboxWorkerFactory func(launcher, workerScript, descriptor, pluginRoot, projectID string, manifest plugin.Manifest, policy plugin.IsolationPolicy) (*plugin.TypeScriptWorker, error)

func sandboxRuntime(platform, local, address string) (string, string, sandboxWorkerFactory, error) {
	if platform == "darwin" {
		if local != "1" || !strings.HasPrefix(address, "127.0.0.1:") {
			return "", "", nil, fmt.Errorf("macOS worker runtime is for loopback-only local development; use make dev server")
		}
		log.Print("macOS local development: Seatbelt worker isolation (not the Linux production sandbox)")
		return "scripts/run-macos-typescript-worker.sh", "/usr/bin/sandbox-exec", plugin.NewManifestMacOSTypeScriptWorker, nil
	}
	return "scripts/run-sandboxed-typescript-worker.sh", "bwrap", plugin.NewManifestBubblewrapTypeScriptWorker, nil
}

func bindSandboxedTypeScriptWorkers(registry *plugin.Registry, locations map[string]plugin.PluginLocation) error {
	launcherPath, sandboxBinary, workerFactory, err := sandboxRuntime(runtime.GOOS, os.Getenv("BRAIN_LOCAL_MACOS"), os.Getenv("HTTP_ADDRESS"))
	if err != nil {
		return err
	}
	launcher, err := filepath.Abs(launcherPath)
	if err != nil {
		return err
	}
	workerScript, err := filepath.Abs("scripts/typescript-worker.mjs")
	if err != nil {
		return err
	}
	for _, binary := range []string{sandboxBinary, "node"} {
		if _, err := exec.LookPath(binary); err != nil {
			return fmt.Errorf("sandbox runtime requires %s: %w", binary, err)
		}
	}
	if _, err := os.Stat("node_modules/tsx/dist/loader.mjs"); err != nil {
		return fmt.Errorf("sandbox TypeScript loader unavailable: %w", err)
	}
	// The address-space limit includes V8's reservations. The launcher also
	// caps the JavaScript heap; CPU limits apply over a worker's lifetime.
	// RLIMIT_NPROC counts threads across the shared host UID, not per worker.
	// Reserve a bounded aggregate budget for the complete plugin inventory.
	policy := plugin.IsolationPolicy{FilesystemReadOnly: true, NetworkEnabled: false, CPUMillis: 30000, MemoryBytes: 2 << 30, ProcessLimit: 512, WallTime: 2 * time.Second}
	probeRoot, err := os.MkdirTemp("", "brain-sandbox-probe-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(probeRoot)
	probeSource := []byte("export function health(): { accepted: boolean } { const memory = new WebAssembly.Memory({ initial: 1, maximum: 2 }); return { accepted: memory.buffer.byteLength === 65536 }; }\n")
	if err := os.WriteFile(filepath.Join(probeRoot, "index.mts"), probeSource, 0o600); err != nil {
		return err
	}
	probeManifest := plugin.Manifest{ID: "sandbox-probe", Version: "1.0.0", HostAPI: 1, Entry: "index.mts", SourceFiles: []string{"index.mts"}, Capabilities: []string{"health"}, Routes: []plugin.Route{{Operation: "health", Function: "health", Capability: "health"}}}
	probe, err := workerFactory(launcher, workerScript, filepath.Join(probeRoot, "execution.json"), probeRoot, "sandbox-probe", probeManifest, policy)
	if err != nil {
		return fmt.Errorf("sandbox probe configuration failed: %w", err)
	}
	// Only the host-generated probe is allowed to emit startup diagnostics.
	probe.SetTrustedStartupDiagnostics(os.Stderr)
	if err := probe.Start(); err != nil {
		return fmt.Errorf("sandbox startup probe failed: %w", err)
	}
	response, err := plugin.InvokeForProject(probe, plugin.WorkerRequest{Protocol: 1, ProjectID: "sandbox-probe", PluginID: probeManifest.ID, Version: probeManifest.Version, Operation: "health"}, 1, "sandbox-probe")
	_ = probe.Stop()
	if err != nil || !response.Success || string(response.Payload) != `{"accepted":true}` {
		return fmt.Errorf("sandbox invocation probe failed")
	}
	registry.BindWorkerFactory(func(projectID, pluginID, version string) (plugin.ContractWorker, error) {
		location, ok := locations[pluginID+"@"+version]
		if !ok {
			return nil, fmt.Errorf("plugin location unavailable: %s@%s", pluginID, version)
		}
		manifest, err := plugin.LoadManifest(location.ManifestPath)
		if err != nil {
			return nil, err
		}
		root, err := filepath.Abs(location.Directory)
		if err != nil {
			return nil, err
		}
		root, err = filepath.EvalSymlinks(root)
		if err != nil {
			return nil, err
		}
		descriptor, err := os.CreateTemp("", "brain-worker-*.json")
		if err != nil {
			return nil, err
		}
		_ = descriptor.Close()
		worker, err := workerFactory(launcher, workerScript, descriptor.Name(), root, projectID, manifest, policy)
		if err != nil {
			_ = os.Remove(descriptor.Name())
			return nil, err
		}
		return &temporaryDescriptorWorker{TypeScriptWorker: worker, descriptor: descriptor.Name()}, nil
	})
	return nil
}

// The descriptor is needed only until the sandbox acknowledges module startup.
type temporaryDescriptorWorker struct {
	*plugin.TypeScriptWorker
	descriptor string
}

func (w *temporaryDescriptorWorker) Start() error {
	defer os.Remove(w.descriptor)
	return w.TypeScriptWorker.Start()
}

func verifyObjectManifest(path string) {
	if _, err := deploy.ReadObjectVersionManifest(path); err != nil {
		log.Fatal(err)
	}
	log.Printf("object manifest verified: %s", path)
}

func bootstrapDashboard() {
	bootstrap, err := app.BootstrapDashboard()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("dashboard plugin %s@%s ready at %s\n", bootstrap.PluginID, bootstrap.Version, bootstrap.EntryPoint)
}

func bootstrapOperator() {
	username := os.Getenv("OPERATOR_USERNAME")
	secret := os.Getenv("OPERATOR_SECRET")
	store := &operator.Store{}
	_, err := store.Bootstrap(username, secret, func(value string) string {
		digest := sha256.Sum256([]byte(value))
		return hex.EncodeToString(digest[:])
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("operator %q bootstrapped", username)
}
