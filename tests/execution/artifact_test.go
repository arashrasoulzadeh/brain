package execution_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"brain/internal/execution"
)

func TestBundleCanonicalAndVerified(t *testing.T) {
	files := []execution.BundleFile{{Path: "plugin.yaml", Data: []byte("version: 1")}, {Path: "functions/run.ts", Data: []byte("export {}")}}
	bundle, err := execution.BuildBundle(files, 4096)
	if err != nil {
		t.Fatal(err)
	}
	other, err := execution.BuildBundle([]execution.BundleFile{files[1], files[0]}, 4096)
	if err != nil || !bytes.Equal(bundle, other) {
		t.Fatal("bundle depends on input order")
	}
	a := execution.Artifact{Digest: execution.ArtifactDigest(bundle)}
	files[0].Data[0] = 'X'
	opened, err := execution.OpenBundle(a, bundle, 4096)
	if err != nil || len(opened) != 2 || opened[0].Path != "functions/run.ts" || string(opened[1].Data) != "version: 1" {
		t.Fatalf("bad snapshot: %v %v", opened, err)
	}
	opened[1].Data[0] = 'Y'
	again, err := execution.OpenBundle(a, bundle, 4096)
	if err != nil || string(again[1].Data) != "version: 1" {
		t.Fatal("opened bytes alias the bundle")
	}
	bundle[0] ^= 1
	if _, err := execution.OpenBundle(a, bundle, 4096); !errors.Is(err, execution.ErrArtifactMismatch) {
		t.Fatal("tampering accepted")
	}
}

func TestBundleRejectsUnsafeEntriesAndLimits(t *testing.T) {
	for _, path := range []string{"", ".", "../a", "/a", "a/../b", "a//b", "a\\b", "C:a", "a\x00b", "é.ts", strings.Repeat("a/", 32) + "b"} {
		t.Run(path, func(t *testing.T) {
			if _, err := execution.BuildBundle([]execution.BundleFile{{Path: path}}, 4096); !errors.Is(err, execution.ErrInvalidBundle) {
				t.Fatal("unsafe path accepted")
			}
		})
	}
	for _, paths := range [][]string{{"a", "a"}, {"a", "A"}, {"a", "a/b"}, {"A", "a/b"}, {"A/x", "a/y"}} {
		if _, err := execution.BuildBundle([]execution.BundleFile{{Path: paths[0]}, {Path: paths[1]}}, 4096); !errors.Is(err, execution.ErrInvalidBundle) {
			t.Fatalf("colliding paths accepted: %v", paths)
		}
	}
	if _, err := execution.BuildBundle(nil, 4096); !errors.Is(err, execution.ErrInvalidBundle) {
		t.Fatal("empty bundle accepted")
	}
	for _, max := range []int64{0, 1, 2, 10, execution.MaxBundleBytes + 1} {
		if _, err := execution.BuildBundle([]execution.BundleFile{{Path: "a", Data: []byte("abc")}}, max); !errors.Is(err, execution.ErrInvalidBundle) {
			t.Fatalf("invalid bound accepted: %d", max)
		}
	}
	if _, err := execution.BuildBundle(make([]execution.BundleFile, execution.MaxBundleFiles+1), execution.MaxBundleBytes); !errors.Is(err, execution.ErrInvalidBundle) {
		t.Fatal("too many files accepted")
	}
}

func TestBundleRejectsNonCanonicalEncoding(t *testing.T) {
	for _, raw := range []string{
		`not json`,
		`{"format":1,"files":null}`,
		`{"format":1,"files":[not json]}`,
		`{"format":1,"files":[{"path":"a","data":""}}`,
		`{"format":1,"files":[]}`,
		`{"format":1,"files":[` + strings.Repeat(`{},`, execution.MaxBundleFiles) + `{}]}`,

		`{"format":2,"files":[{"path":"a","data":""}]}`,
		`{"format":1,"files":[{"path":"../a","data":""}]}`,
		`{"format":1,"files":[{"path":"a","data":""}]} `,
		`{"format":1,"files":[{"path":"a","data":null}]}`,
		`{"format":1,"files":[{"path":"a","data":"","extra":true}]}`,
	} {
		bundle := []byte(raw)
		a := execution.Artifact{Digest: execution.ArtifactDigest(bundle)}
		if _, err := execution.OpenBundle(a, bundle, execution.MaxBundleBytes); !errors.Is(err, execution.ErrInvalidBundle) {
			t.Fatalf("noncanonical bundle accepted: %s", raw)
		}
	}
	bundle, err := execution.BuildBundle([]execution.BundleFile{{Path: "a"}}, 4096)
	if err != nil {
		t.Fatal(err)
	}
	a := execution.Artifact{Digest: execution.ArtifactDigest(bundle)}
	if _, err := execution.OpenBundle(a, bundle, 1); !errors.Is(err, execution.ErrInvalidBundle) {
		t.Fatal("oversized bundle accepted")
	}
}
