package plugin

import (
	"database/sql"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileMetadataMigrationAndPersistence(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "files.sqlite")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TABLE plugin_file_rows (project_id TEXT NOT NULL, file_key TEXT NOT NULL, data TEXT NOT NULL, published INTEGER NOT NULL, PRIMARY KEY(project_id,file_key)); INSERT INTO plugin_file_rows VALUES ('one','legacy.txt','legacy',1)`); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	storage, err := OpenSQLiteStorage(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	data, published, legacy, err := storage.fileWithMetadata("one", "legacy.txt")
	if err != nil || data != "legacy" || !published || legacy.Extension != "txt" || legacy.SizeBytes != 6 || legacy.CreatedAt != "" {
		t.Fatalf("legacy migration: %q %t %#v %v", data, published, legacy, err)
	}
	request := WorkerRequest{ProjectID: "one", PluginID: "files", Operation: "files.upload", Payload: []byte(`{"key":"custom-key","data":"data:application/octet-stream;base64,AP+A","filename":"folder/report.PNG","contentType":"image/png","metadata":{"size_bytes":999,"sha256":"forged"}}`)}
	prepared, commit, err := prepareLegacyStorageRequest(storage, request)
	if err != nil {
		t.Fatal(err)
	}
	if commit == nil {
		t.Fatal("no persistence effect")
	}
	var payload struct {
		Metadata FileMetadata `json:"metadata"`
	}
	if err := json.Unmarshal(prepared.Payload, &payload); err != nil {
		t.Fatal(err)
	}
	meta := payload.Metadata
	if meta.Filename != "report.PNG" || meta.Extension != "png" || meta.SizeBytes != 3 || meta.Encoding != "base64" || meta.SHA256 == "forged" || len(meta.SHA256) != 64 || meta.DeclaredContentType != "image/png" || meta.ContentType != "application/octet-stream" || meta.CreatedAt == "" {
		t.Fatalf("wrong metadata: %#v", meta)
	}
	if err := commit(); err != nil {
		t.Fatal(err)
	}
	request.Operation = "files.publish"
	request.Payload = []byte(`{"key":"custom-key","metadata":{"filename":"forged"}}`)
	_, publish, err := prepareLegacyStorageRequest(storage, request)
	if err != nil {
		t.Fatal(err)
	}
	if err := publish(); err != nil {
		t.Fatal(err)
	}
	if err := storage.Close(); err != nil {
		t.Fatal(err)
	}
	storage, err = OpenSQLiteStorage(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	_, published, persisted, err := storage.fileWithMetadata("one", "custom-key")
	if err != nil || !published || persisted.Filename != meta.Filename || persisted.SHA256 != meta.SHA256 || persisted.CreatedAt != meta.CreatedAt || persisted.PublishedAt == "" {
		t.Fatalf("metadata lost after publish/restart: %#v %v", persisted, err)
	}
	if data, _, _, err := storage.fileWithMetadata("two", "custom-key"); err != nil || data != "" {
		t.Fatal("cross-project file leaked")
	}
	page, err := (legacyStoragePlanner{storage: storage}).AdminInventory("one", "files", "file_keys", "files.metadata", "", 10)
	if err != nil || len(page.Rows) != 2 || page.Rows[0].Filename != "report.PNG" || *page.Rows[0].SizeBytes != 3 {
		t.Fatalf("incorrect inventory: %#v %v", page, err)
	}
	request.Operation = "files.upload"
	request.Payload = []byte(`{"key":"custom-key","data":"replacement","filename":"new.txt"}`)
	_, overwrite, err := prepareLegacyStorageRequest(storage, request)
	if err != nil {
		t.Fatal(err)
	}
	if err := overwrite(); err != nil {
		t.Fatal(err)
	}
	_, published, updated, err := storage.fileWithMetadata("one", "custom-key")
	if err != nil || published || updated.Filename != "new.txt" || updated.CreatedAt != meta.CreatedAt || updated.PublishedAt != "" || updated.SHA256 == meta.SHA256 {
		t.Fatalf("overwrite metadata: %#v %v", updated, err)
	}
	if strings.Contains(safeFilename("../bad\x00.txt"), "\x00") {
		t.Fatal("unsafe filename")
	}
}

func TestPublicConfigurationReadbackIsExplicitAndProjectScoped(t *testing.T) {
	r := NewRegistry(1)
	m := Manifest{ID: "custom", Version: "1", HostAPI: 1, ConfigSchema: map[string]string{"count": "int", "secret": "string"}, PublicConfig: []string{"count"}}
	if err := r.Register(m); err != nil {
		t.Fatal(err)
	}
	for _, project := range []string{"one", "two"} {
		if err := r.Install(project, m.ID, m.Version); err != nil {
			t.Fatal(err)
		}
	}
	if err := r.Configure("one", m.ID, m.Version, map[string]any{"count": 5, "secret": "private-value"}); err != nil {
		t.Fatal(err)
	}
	one, err := r.Status("one", m.ID, m.Version)
	if err != nil || one.PublicConfig["count"] != 5 || len(one.PublicConfig) != 1 {
		t.Fatalf("wrong public config: %#v %v", one.PublicConfig, err)
	}
	two, err := r.Status("two", m.ID, m.Version)
	if err != nil || len(two.PublicConfig) != 0 {
		t.Fatal("configuration leaked between projects")
	}
	encoded, err := json.Marshal(one)
	if err != nil || strings.Contains(string(encoded), "private-value") {
		t.Fatal("secret value exposed")
	}
	m.ID = "bad"
	m.PublicConfig = []string{"undeclared"}
	if err := r.Register(m); err == nil {
		t.Fatal("undeclared readback accepted")
	}
}
