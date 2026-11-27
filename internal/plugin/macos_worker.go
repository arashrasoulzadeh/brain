package plugin

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// NewManifestMacOSTypeScriptWorker is a local-development adapter, not a
// substitute for Linux namespaces or production resource enforcement.
func NewManifestMacOSTypeScriptWorker(launcher, workerScript, descriptor, pluginRoot, projectID string, manifest Manifest, policy IsolationPolicy) (*TypeScriptWorker, error) {
	if runtime.GOOS != "darwin" {
		return nil, ErrInvalidIsolationPolicy
	}
	if err := policy.Validate(); err != nil {
		return nil, err
	}
	node, err := exec.LookPath("node")
	if err != nil {
		return nil, err
	}
	node, err = filepath.EvalSymlinks(node)
	if err != nil {
		return nil, err
	}
	// Seatbelt filters canonical paths. In particular, macOS temporary paths
	// can use /var aliases to /private/var.
	workerScript, err = filepath.EvalSymlinks(workerScript)
	if err != nil {
		return nil, err
	}
	pluginRoot, err = filepath.EvalSymlinks(pluginRoot)
	if err != nil {
		return nil, err
	}
	worker, err := NewManifestBubblewrapTypeScriptWorker(launcher, workerScript, descriptor, pluginRoot, projectID, manifest, policy)
	if err != nil {
		return nil, err
	}
	descriptorPath, err := filepath.EvalSymlinks(descriptor)
	if err != nil {
		return nil, err
	}
	root := filepath.Dir(filepath.Dir(workerScript))
	profile := strings.Builder{}
	profile.WriteString(`(version 1)
(deny default)
(allow process-fork)
(allow sysctl-read)
(allow file-read-metadata)
(allow mach-lookup (global-name "com.apple.system.logger") (global-name "com.apple.system.notification_center") (global-name "com.apple.securityd") (global-name "com.apple.SecurityServer") (global-name "com.apple.cfprefsd.agent") (global-name "com.apple.cfprefsd.daemon"))
(allow file-read* (subpath "/System/Library") (subpath "/usr/lib") (subpath "/usr/share") (literal "/dev/null") (literal "/dev/random") (literal "/dev/urandom") (literal "/private/etc/localtime") (literal "/private/etc/hosts") (literal "/private/etc/resolv.conf"))
(allow file-write-data (literal "/dev/null"))
`)
	// The package manager's installed libraries are read-only, not the
	// user's home or arbitrary project directories.
	for _, directory := range []string{"/opt/homebrew/Cellar", "/opt/homebrew/opt", "/usr/local/Cellar", "/usr/local/opt", "/usr/local/lib"} {
		if info, err := os.Stat(directory); err == nil && info.IsDir() {
			fmt.Fprintf(&profile, "(allow file-read* (subpath %s))\n", strconv.Quote(directory))
		}
	}
	for _, path := range []string{node, workerScript, descriptorPath, filepath.Join(root, "package.json"), filepath.Join(root, "tsconfig.json")} {
		fmt.Fprintf(&profile, "(allow file-read* (literal %s))\n", strconv.Quote(path))
	}
	for _, directory := range []string{filepath.Join(root, "node_modules"), filepath.Join(root, "plugins", "sdk-typescript")} {
		fmt.Fprintf(&profile, "(allow file-read* (subpath %s))\n", strconv.Quote(directory))
	}
	for _, source := range manifest.SourceFiles {
		path := filepath.Join(pluginRoot, source)
		canonical, err := filepath.EvalSymlinks(path)
		if err != nil || canonical != path {
			return nil, ErrInvalidManifest
		}
		fmt.Fprintf(&profile, "(allow file-read* (literal %s))\n", strconv.Quote(path))
	}
	fmt.Fprintf(&profile, "(allow process-exec (literal %s))\n", strconv.Quote(node))
	// tsx uses the installed esbuild service. Allow only the packaged native
	// executables, not arbitrary binaries or a shell.
	for _, pattern := range []string{"node_modules/@esbuild/darwin-*/bin/esbuild", "node_modules/esbuild/bin/esbuild"} {
		paths, err := filepath.Glob(filepath.Join(root, pattern))
		if err != nil {
			return nil, err
		}
		for _, path := range paths {
			canonical, err := filepath.EvalSymlinks(path)
			if err != nil {
				return nil, err
			}
			if !strings.HasPrefix(canonical, filepath.Join(root, "node_modules")+string(filepath.Separator)) {
				return nil, ErrInvalidManifest
			}
			fmt.Fprintf(&profile, "(allow process-exec (literal %s))\n", strconv.Quote(canonical))
		}
	}
	worker.env = append(worker.env, "NODE_BIN="+node, "BRAIN_MACOS_PROFILE="+profile.String())
	return worker, nil
}
