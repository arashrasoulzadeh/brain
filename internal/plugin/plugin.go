package plugin

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrDuplicate          = errors.New("plugin version already registered")
	ErrIncompatible       = errors.New("plugin host API incompatible")
	ErrNotRegistered      = errors.New("plugin version not registered")
	ErrNotInstalled       = errors.New("plugin not installed for project")
	ErrDisabled           = errors.New("plugin disabled")
	ErrGrantDenied        = errors.New("plugin grant denied")
	ErrProjectInvalid     = errors.New("project identity invalid")
	ErrDependencyMissing  = errors.New("plugin dependency missing")
	ErrInvalidManifest    = errors.New("invalid plugin manifest")
	ErrInvalidConfig      = errors.New("invalid plugin configuration")
	ErrSchemaIncompatible = errors.New("plugin configuration schema incompatible")
	ErrNotReady           = errors.New("plugin worker not ready")
	ErrVersionReferenced  = errors.New("plugin version still referenced")
	ErrCapabilityApproval = errors.New("new plugin capability requires approval")
	ErrStalePreview       = errors.New("plugin lifecycle preview is stale")
	ErrActiveWork         = errors.New("project has active plugin work")
)

type Registered struct{ Manifest Manifest }

type Installation struct {
	ProjectID  string
	PluginID   string
	Version    string
	Enabled    bool
	Grants     map[string]bool
	Config     map[string]any
	Resources  map[string]map[string]bool
	Draining   bool
	DrainUntil time.Time
}

type InstallationStatus struct {
	ProjectID   string
	PluginID    string
	Version     string
	Enabled     bool
	Ready       bool
	WorkerReady bool
	Grants      []string
	Draining    bool
	DrainUntil  time.Time
}

type Registry struct {
	HostAPI             int
	registered          map[string]Registered
	installations       map[string]Installation
	installationMu      sync.RWMutex
	readiness           map[string]bool
	readinessMu         sync.RWMutex
	grantMu             sync.RWMutex
	approvalMu          sync.RWMutex
	audit               []AuditEvent
	artifactMu          sync.RWMutex
	auditMu             sync.RWMutex
	active              map[string]string
	pins                map[string]string
	artifacts           map[string]string
	approvals           map[string]bool
	lifecycleRevision   uint64
	lifecycleMu         sync.RWMutex
	durableData         map[string][]byte
	durableMu           sync.RWMutex
	supervisor          *Supervisor
	workerMu            sync.RWMutex
	workers             map[string]ContractWorker
	workerFactory       func(projectID, pluginID, version string) (ContractWorker, error)
	storage             *SQLiteStorage
	storageInvocations  [64]sync.Mutex
	webSockets          *WebSocketRegistry
}

func NewRegistry(hostAPI int) *Registry {
	return &Registry{HostAPI: hostAPI, registered: map[string]Registered{}, installations: map[string]Installation{}, readiness: map[string]bool{}, active: map[string]string{}, pins: map[string]string{}, artifacts: map[string]string{}, approvals: map[string]bool{}, durableData: map[string][]byte{}, workers: map[string]ContractWorker{}}
}
