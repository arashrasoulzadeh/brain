package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
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
		log.Fatal(http.ListenAndServe(cfg.HTTPAddress, app.NewServer().Handler()))
	}
	server, err := productionServer()
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(http.ListenAndServe(cfg.HTTPAddress, server.Handler()))
}

func productionServer() (*app.Server, error) {
	username, secret := os.Getenv("OPERATOR_USERNAME"), os.Getenv("OPERATOR_SECRET")
	if username == "" || secret == "" {
		return nil, fmt.Errorf("OPERATOR_USERNAME and OPERATOR_SECRET are required")
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
	bindLocalTypeScriptWorkers(registry, locations)
	projectsPath := os.Getenv("PROJECTS_FILE")
	if projectsPath == "" {
		projectsPath = ".local/brain.sqlite"
	}
	if err := os.MkdirAll(filepath.Dir(projectsPath), 0o700); err != nil {
		return nil, fmt.Errorf("project storage directory unavailable: %w", err)
	}
	projects, err := project.NewSQLiteStore(projectsPath)
	if err != nil {
		return nil, fmt.Errorf("project storage unavailable: %w", err)
	}
	storage, err := plugin.OpenSQLiteStorage(projectsPath)
	if err != nil {
		_ = projects.Close()
		return nil, fmt.Errorf("plugin storage unavailable: %w", err)
	}
	if err := registry.BindSQLiteStorage(storage); err != nil {
		_ = storage.Close()
		_ = projects.Close()
		return nil, fmt.Errorf("registry state unavailable: %w", err)
	}
	server := app.NewServerWithOperatorAdmin(admin.NewLifecycleServiceWithProjects(operators, registry, projects), operators, digest)
	server.SetReadinessCheck(func(ctx context.Context) error {
		if err := projects.CheckHealth(ctx); err != nil {
			return err
		}
		if err := storage.CheckHealth(ctx); err != nil {
			return err
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

func bindLocalTypeScriptWorkers(registry *plugin.Registry, locations map[string]plugin.PluginLocation) {
	command := filepath.Join("node_modules", ".bin", "tsx")
	workerScript := filepath.Join("scripts", "typescript-worker.mjs")
	policy := plugin.IsolationPolicy{FilesystemReadOnly: true, NetworkEnabled: false, CPUMillis: 1000, MemoryBytes: 128 << 20, ProcessLimit: 16, WallTime: 2 * time.Second}
	_ = os.MkdirAll(".local", 0o700)
	registry.BindWorkerFactory(func(projectID, pluginID, version string) (plugin.ContractWorker, error) {
		location, ok := locations[pluginID+"@"+version]
		if !ok {
			return nil, fmt.Errorf("plugin location unavailable: %s@%s", pluginID, version)
		}
		manifest, err := plugin.LoadManifest(location.ManifestPath)
		if err != nil {
			return nil, err
		}
		descriptor, err := os.CreateTemp(".local", "worker-*.json")
		if err != nil {
			return nil, err
		}
		_ = descriptor.Close()
		return plugin.NewManifestTypeScriptWorker(command, workerScript, descriptor.Name(), location.Directory, projectID, manifest, policy)
	})
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
