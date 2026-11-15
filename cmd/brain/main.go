package main

import (
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
	server := productionServer()
	log.Fatal(http.ListenAndServe(cfg.HTTPAddress, server.Handler()))
}

func productionServer() *app.Server {
	username, secret := os.Getenv("OPERATOR_USERNAME"), os.Getenv("OPERATOR_SECRET")
	if username == "" || secret == "" {
		return app.NewServer()
	}
	operators := &operator.Store{}
	digest := func(value string) string {
		sum := sha256.Sum256([]byte(value))
		return hex.EncodeToString(sum[:])
	}
	if _, err := operators.Bootstrap(username, secret, digest); err != nil {
		log.Printf("operator admin disabled: %v", err)
		return app.NewServer()
	}
	registry := plugin.NewRegistry(1)
	locations := registerDiscoveredPlugins(registry)
	bindLocalTypeScriptWorkers(registry, locations)
	projectsPath := os.Getenv("PROJECTS_FILE")
	if projectsPath == "" {
		projectsPath = ".local/brain.sqlite"
	}
	projects, err := project.NewSQLiteStore(projectsPath)
	if err != nil {
		log.Printf("project store unavailable: %v", err)
		projects = project.NewStore()
	}
	return app.NewServerWithOperatorAdmin(admin.NewLifecycleServiceWithProjects(operators, registry, projects), operators, digest)
}

func registerDiscoveredPlugins(registry *plugin.Registry) map[string]plugin.PluginLocation {
	registered := map[string]plugin.PluginLocation{}
	locations, err := plugin.Discover("plugins")
	if err != nil {
		log.Printf("plugin discovery unavailable: %v", err)
		return registered
	}
	for _, location := range locations {
		manifest, err := plugin.LoadManifest(location.ManifestPath)
		if err != nil {
			log.Printf("plugin %s rejected: %v", location.ID, err)
			continue
		}
		if err := registry.Register(manifest); err != nil {
			log.Printf("plugin %s@%s rejected: %v", manifest.ID, manifest.Version, err)
			continue
		}
		registered[manifest.ID+"@"+manifest.Version] = location
	}
	return registered
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
