package execution

import (
	"bytes"
	"errors"
	"slices"
	"sync"
)

var (
	ErrArtifactNotFound = errors.New("execution artifact not registered")
	ErrArtifactConflict = errors.New("execution artifact version already pinned")
	ErrCatalogLimit     = errors.New("execution artifact catalog capacity exceeded")
)

const maxArtifactEntrypoints = 1024

// ArtifactKey is an exact registry identity. An empty ProjectID explicitly
// selects a shared artifact; there is no fallback between scopes.
type ArtifactKey struct {
	ProjectID, ID, Version string
}

type Registration struct {
	Key         ArtifactKey
	Runtime     Runtime
	Entrypoints []string
	Bundle      []byte
}

type catalogEntry struct {
	digest      string
	runtime     Runtime
	entrypoints []string
	bundle      []byte
}

// Catalog is a bounded, in-memory immutable artifact index. Only trusted host
// code may register or resolve identities. It is not a permission broker or a
// durable registry; production must persist and restore version pins separately.
type Catalog struct {
	mu         sync.RWMutex
	entries    map[ArtifactKey]catalogEntry
	maxEntries int
	maxBytes   int64
	usedBytes  int64
}

func NewCatalog(maxEntries int, maxBytes int64) (*Catalog, error) {
	if maxEntries <= 0 || maxBytes <= 0 {
		return nil, ErrCatalogLimit
	}
	return &Catalog{entries: make(map[ArtifactKey]catalogEntry), maxEntries: maxEntries, maxBytes: maxBytes}, nil
}

// Register pins content, entrypoints and runtime requirements to one version.
// A repeat of the same registration is harmless; a change requires a new version.
func (c *Catalog) Register(reg Registration) error {
	if !present(reg.Key.ID) || !validVersion(reg.Key.Version) || (reg.Key.ProjectID != "" && !present(reg.Key.ProjectID)) ||
		!present(reg.Runtime.Kind) || !present(reg.Runtime.Profile) || len(reg.Entrypoints) == 0 || len(reg.Entrypoints) > maxArtifactEntrypoints {
		return ErrInvalidRequest
	}
	entrypoints := slices.Clone(reg.Entrypoints)
	slices.Sort(entrypoints)
	for i, entrypoint := range entrypoints {
		if !present(entrypoint) || (i > 0 && entrypoint == entrypoints[i-1]) {
			return ErrInvalidRequest
		}
	}
	if int64(len(reg.Bundle)) > min(c.maxBytes, MaxBundleBytes) {
		return ErrCatalogLimit
	}
	bundle := bytes.Clone(reg.Bundle)
	digest := ArtifactDigest(bundle)
	if _, err := OpenBundle(Artifact{Digest: digest}, bundle, min(c.maxBytes, MaxBundleBytes)); err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if existing, ok := c.entries[reg.Key]; ok {
		if existing.digest == digest && existing.runtime == reg.Runtime && slices.Equal(existing.entrypoints, entrypoints) {
			return nil
		}
		return ErrArtifactConflict
	}
	if len(c.entries) >= c.maxEntries || int64(len(bundle)) > c.maxBytes-c.usedBytes {
		return ErrCatalogLimit
	}
	c.entries[reg.Key] = catalogEntry{digest: digest, runtime: reg.Runtime, entrypoints: entrypoints, bundle: bundle}
	c.usedBytes += int64(len(bundle))
	return nil
}

// Resolve selects an already-authorized exact key and declared entrypoint.
func (c *Catalog) Resolve(key ArtifactKey, entrypoint string) (Artifact, Runtime, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, ok := c.entries[key]
	if !ok || !slices.Contains(entry.entrypoints, entrypoint) {
		return Artifact{}, Runtime{}, ErrArtifactNotFound
	}
	return Artifact{ProjectID: key.ProjectID, ID: key.ID, Version: key.Version, Digest: entry.digest, Entrypoint: entrypoint}, entry.runtime, nil
}

// Fetch retrieves a copy of the exact pinned bytes. Brain must independently
// authorize each artifact fetch for its current assigned execution and worker.
func (c *Catalog) Fetch(artifact Artifact) ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, ok := c.entries[ArtifactKey{ProjectID: artifact.ProjectID, ID: artifact.ID, Version: artifact.Version}]
	if !ok || !slices.Contains(entry.entrypoints, artifact.Entrypoint) {
		return nil, ErrArtifactNotFound
	}
	if artifact.Digest != entry.digest {
		return nil, ErrArtifactMismatch
	}
	return bytes.Clone(entry.bundle), nil
}
