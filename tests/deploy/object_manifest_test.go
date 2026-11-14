package deploy_test

import (
	"errors"
	"testing"

	"brain/internal/deploy"
)

func TestObjectManifestRecordsProjectScopedVersions(t *testing.T) {
	path := t.TempDir() + "/objects.json"
	manifest := deploy.ObjectVersionManifest{Objects: []deploy.ObjectVersion{
		{ProjectID: "dev", ObjectKey: "uploads/avatar", VersionID: "v1"},
		{ProjectID: "dev", ObjectKey: "uploads/logo", VersionID: "v2"},
	}}
	if err := deploy.RecordObjectVersionManifest(path, manifest); err != nil {
		t.Fatal(err)
	}
	read, err := deploy.ReadObjectVersionManifest(path)
	if err != nil || len(read.Objects) != 2 {
		t.Fatalf("manifest = %#v, err = %v", read, err)
	}
	manifest.Objects = append(manifest.Objects, manifest.Objects[0])
	if err := deploy.RecordObjectVersionManifest(path, manifest); !errors.Is(err, deploy.ErrInvalidObjectVersionManifest) {
		t.Fatal("duplicate object version was accepted")
	}
	available := map[string]bool{"dev\x00uploads/avatar\x00v1": true}
	missing := deploy.MissingObjectVersions(manifest.Objects[:2], available)
	if len(missing) != 1 || missing[0].ObjectKey != "uploads/logo" {
		t.Fatalf("missing objects = %#v", missing)
	}
}
