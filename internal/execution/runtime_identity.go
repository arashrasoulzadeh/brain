package execution

import (
	"io/fs"
	"os"
	"path/filepath"
)

// RuntimeSourceDigest compares the trusted runtime distribution used by Brain
// and an enrolled Linux Neuron. It is not an attestation of the remote machine
// or of installed Node dependencies; operators must deploy a verified image.
func RuntimeSourceDigest(root string) (string, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		return "", err
	}
	var files []BundleFile
	remaining := MaxBundleBytes
	add := func(name string) error {
		info, err := os.Lstat(name)
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() || len(files) >= MaxBundleFiles || info.Size() > remaining {
			return ErrInvalidBundle
		}
		relative, err := filepath.Rel(root, name)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(name)
		if err != nil {
			return err
		}
		remaining -= int64(len(data))
		if remaining < 0 {
			return ErrInvalidBundle
		}
		files = append(files, BundleFile{Path: filepath.ToSlash(relative), Data: data})
		return nil
	}
	for _, name := range []string{"package.json", "package-lock.json", "tsconfig.json", "scripts/typescript-worker.mjs", "scripts/run-sandboxed-typescript-worker.sh"} {
		if err := add(filepath.Join(root, name)); err != nil {
			return "", err
		}
	}
	sdk := filepath.Join(root, "plugins", "sdk-typescript")
	if err := filepath.WalkDir(sdk, func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		return add(name)
	}); err != nil {
		return "", err
	}
	// The package entrypoint must exist even if the SDK directory is empty.
	if info, err := os.Lstat(filepath.Join(sdk, "src", "index.ts")); err != nil || !info.Mode().IsRegular() {
		return "", ErrInvalidBundle
	}
	bundle, err := BuildBundle(files, MaxBundleBytes)
	if err != nil {
		return "", err
	}
	return ArtifactDigest(bundle), nil
}
