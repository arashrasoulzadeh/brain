package plugin

import (
	"os"
	"path/filepath"
	"strings"
)

// BindBundledRuntime mounts the trusted SDK and package scope beside a fresh
// artifact source directory, preserving existing relative SDK imports. It
// never exposes another plugin directory or expands the artifact's source list.
// Only the Neuron runtime calls this before starting a sandbox.
func (w *TypeScriptWorker) BindBundledRuntime(pluginRoot string) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	if w.ready || w.command == nil || filepath.Base(pluginRoot) != "source" {
		return ErrInvalidIsolationPolicy
	}
	var runtimeRoot string
	for _, entry := range w.env {
		if value, ok := strings.CutPrefix(entry, "WORKER_RUNTIME_ROOT="); ok {
			runtimeRoot = value
		}
	}
	if runtimeRoot == "" || !filepath.IsAbs(pluginRoot) || !filepath.IsAbs(runtimeRoot) {
		return ErrInvalidIsolationPolicy
	}
	parent := filepath.Dir(pluginRoot)
	for _, link := range []struct {
		target string
		name   string
	}{
		{filepath.Join(runtimeRoot, "plugins", "sdk-typescript"), "sdk-typescript"},
		{filepath.Join(runtimeRoot, "package.json"), "package.json"},
		{filepath.Join(runtimeRoot, "node_modules"), "node_modules"},
	} {
		canonical, err := filepath.EvalSymlinks(link.target)
		if err != nil || canonical != link.target {
			return ErrInvalidIsolationPolicy
		}
		if err := os.Symlink(link.target, filepath.Join(parent, link.name)); err != nil {
			return err
		}
	}
	w.env = append(w.env, "WORKER_BUNDLE_PARENT="+parent)
	return nil
}
