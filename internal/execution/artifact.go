package execution

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/fs"
	"path"
	"regexp"
	"slices"
	"strings"
)

const (
	MaxBundleBytes     int64 = 64 << 20
	MaxBundleFiles           = 4096
	maxBundlePathDepth       = 32
)

var ErrInvalidBundle = errors.New("invalid execution artifact bundle")
var bundlePath = regexp.MustCompile(`^[A-Za-z0-9._/-]+$`)

type BundleFile struct {
	Path string `json:"path"`
	Data []byte `json:"data"`
}

type artifactBundle struct {
	Format int          `json:"format"`
	Files  []BundleFile `json:"files"`
}

// BuildBundle snapshots a host-selected file set into canonical bytes. The
// registry must select the manifest, all declared sources and dependency
// identities; this function does not discover files or authorize artifacts.
func BuildBundle(files []BundleFile, maxBytes int64) ([]byte, error) {
	if maxBytes <= 0 || maxBytes > MaxBundleBytes || len(files) == 0 || len(files) > MaxBundleFiles {
		return nil, ErrInvalidBundle
	}
	remaining := maxBytes
	seen := make(map[string]bool, len(files))
	prefixes := make(map[string]string, len(files))
	canonical := make([]BundleFile, 0, len(files))
	for _, file := range files {
		if len(file.Path) > 1024 || strings.Count(file.Path, "/") >= maxBundlePathDepth || file.Path == "." || !fs.ValidPath(file.Path) || !bundlePath.MatchString(file.Path) || int64(len(file.Path)) > remaining {
			return nil, ErrInvalidBundle
		}
		remaining -= int64(len(file.Path))
		if int64(len(file.Data)) > remaining {
			return nil, ErrInvalidBundle
		}
		remaining -= int64(len(file.Data))
		key := strings.ToLower(file.Path)
		if seen[key] {
			return nil, ErrInvalidBundle
		}
		seen[key] = true
		for current := file.Path; current != "."; current = path.Dir(current) {
			folded := strings.ToLower(current)
			if previous, ok := prefixes[folded]; ok && previous != current {
				return nil, ErrInvalidBundle
			}
			prefixes[folded] = current
		}
		canonical = append(canonical, BundleFile{Path: file.Path, Data: append([]byte{}, file.Data...)})
	}
	for _, file := range canonical {
		for parent := path.Dir(file.Path); parent != "."; parent = path.Dir(parent) {
			if seen[strings.ToLower(parent)] {
				return nil, ErrInvalidBundle
			}
		}
	}
	slices.SortFunc(canonical, func(a, b BundleFile) int { return strings.Compare(a.Path, b.Path) })
	data, err := json.Marshal(artifactBundle{Format: 1, Files: canonical})
	if err != nil || int64(len(data)) > maxBytes {
		return nil, ErrInvalidBundle
	}
	return data, nil
}

// OpenBundle verifies the exact digest before decoding, rejects ambiguous
// encodings and returns independent bytes. It never writes to the filesystem.
func OpenBundle(artifact Artifact, data []byte, maxBytes int64) ([]BundleFile, error) {
	if maxBytes <= 0 || maxBytes > MaxBundleBytes || int64(len(data)) > maxBytes {
		return nil, ErrInvalidBundle
	}
	if err := artifact.Verify(data); err != nil {
		return nil, err
	}
	const prefix = `{"format":1,"files":`
	if !bytes.HasPrefix(data, []byte(prefix)) || !bytes.HasSuffix(data, []byte(`}`)) {
		return nil, ErrInvalidBundle
	}
	// Decode the array incrementally so a malformed bundle cannot allocate
	// an unbounded number of file records before the count check.
	decoder := json.NewDecoder(bytes.NewReader(data[len(prefix) : len(data)-1]))
	if token, err := decoder.Token(); err != nil || token != json.Delim('[') {
		return nil, ErrInvalidBundle
	}
	files := make([]BundleFile, 0)
	for decoder.More() {
		if len(files) == MaxBundleFiles {
			return nil, ErrInvalidBundle
		}
		var file BundleFile
		if err := decoder.Decode(&file); err != nil {
			return nil, ErrInvalidBundle
		}
		files = append(files, file)
	}
	if token, err := decoder.Token(); err != nil || token != json.Delim(']') {
		return nil, ErrInvalidBundle
	}
	canonical, err := BuildBundle(files, maxBytes)
	if err != nil || !bytes.Equal(canonical, data) {
		return nil, ErrInvalidBundle
	}
	return files, nil
}
