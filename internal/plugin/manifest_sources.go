package plugin

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

var sourceImportPattern = regexp.MustCompile(`(?m)^\s*import[^\n"']*(?:from\s*)?["']([^"']+)["']`)
var sourceExportPattern = regexp.MustCompile(`(?m)^\s*export[^\n"']*from\s*["']([^"']+)["']`)
var sourceDynamicImportPattern = regexp.MustCompile(`(?m)\bimport\s*\(\s*["']([^"']+)["']\s*\)`)
var sourceRequirePattern = regexp.MustCompile(`(?m)\brequire\s*\(\s*["']([^"']+)["']\s*\)`)

func validRelativeManifestPath(candidate string) bool {
	if candidate == "" || strings.HasPrefix(candidate, "/") || strings.ContainsAny(candidate, "\\\x00") {
		return false
	}
	clean := path.Clean(candidate)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return false
	}
	return clean == candidate
}

func containsSymlink(root, relativePath string) bool {
	current := root
	for _, component := range strings.Split(relativePath, string(filepath.Separator)) {
		current = filepath.Join(current, component)
		info, err := os.Lstat(current)
		if err != nil {
			return false
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return true
		}
	}
	return false
}

func ValidatePluginSources(directory string, manifest Manifest) error {
	if err := ValidateManifestContract(manifest); err != nil {
		return err
	}
	allowedImports := map[string]bool{}
	for _, importPath := range manifest.Imports {
		allowedImports[importPath] = true
	}
	for _, sourceFile := range manifest.SourceFiles {
		if !validRelativeManifestPath(sourceFile) {
			return ErrInvalidManifest
		}
		path := filepath.Join(directory, sourceFile)
		if containsSymlink(directory, sourceFile) {
			return ErrInvalidManifest
		}
		info, err := os.Lstat(path)
		if err != nil || !info.Mode().IsRegular() {
			return ErrInvalidManifest
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return ErrInvalidManifest
		}
		matches := sourceImportPattern.FindAllStringSubmatch(string(data), -1)
		matches = append(matches, sourceExportPattern.FindAllStringSubmatch(string(data), -1)...)
		matches = append(matches, sourceDynamicImportPattern.FindAllStringSubmatch(string(data), -1)...)
		matches = append(matches, sourceRequirePattern.FindAllStringSubmatch(string(data), -1)...)
		for _, match := range matches {
			if !allowedImports[match[1]] {
				return fmt.Errorf("%w: %s imports %s", ErrInvalidManifest, sourceFile, match[1])
			}
		}
	}
	return nil
}
