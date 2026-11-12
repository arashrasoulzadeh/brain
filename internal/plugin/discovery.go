package plugin

import (
	"errors"
	"os"
	"path/filepath"
)

var ErrPluginManifestMissing = errors.New("plugin manifest missing")

type PluginLocation struct {
	ID           string
	Directory    string
	ManifestPath string
}

func Discover(root string) ([]PluginLocation, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	locations := make([]PluginLocation, 0)
	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == "sdk-typescript" || entry.Name() == "template" {
			continue
		}
		manifest := filepath.Join(root, entry.Name(), "plugin.yaml")
		if _, err := os.Stat(manifest); err != nil {
			if os.IsNotExist(err) {
				return nil, ErrPluginManifestMissing
			}
			return nil, err
		}
		locations = append(locations, PluginLocation{ID: entry.Name(), Directory: filepath.Join(root, entry.Name()), ManifestPath: manifest})
	}
	return locations, nil
}
