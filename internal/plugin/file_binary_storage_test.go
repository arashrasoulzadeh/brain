package plugin

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"testing/iotest"
)

type binaryTestWriter func([]byte) (int, error)

func (w binaryTestWriter) Write(p []byte) (int, error) { return w(p) }

func binaryTestRegistry(t *testing.T) (*Registry, *SQLiteStorage, string) {
	t.Helper()
	s := coverageStorage(t)
	m, err := LoadManifest("../../plugins/files/plugin.yaml")
	if err != nil {
		t.Fatal(err)
	}
	r := NewRegistry(1)
	if err := r.Register(m); err != nil {
		t.Fatal(err)
	}
	if err := r.BindSQLiteStorage(s); err != nil {
		t.Fatal(err)
	}
	if err := r.SetReadiness(m.ID, m.Version, true); err != nil {
		t.Fatal(err)
	}
	for _, project := range []string{"alpha", "beta"} {
		if err := r.Install(project, m.ID, m.Version); err != nil {
			t.Fatal(err)
		}
		for _, capability := range m.Capabilities {
			if err := r.Grant(project, m.ID, m.Version, capability); err != nil {
				t.Fatal(err)
			}
		}
		if err := r.Configure(project, m.ID, m.Version, map[string]any{"maxUploadBytes": 2 * fileChunkBytes}); err != nil {
			t.Fatal(err)
		}
		if err := r.Enable(project, m.ID, m.Version); err != nil {
			t.Fatal(err)
		}
		if err := r.BindWorker(project, m.ID, m.Version, &confirmedInstallWorker{ready: true}); err != nil {
			t.Fatal(err)
		}
	}
	return r, s, m.Version
}

func TestBinaryFileRoundTripAndRollback(t *testing.T) {
	r, s, version := binaryTestRegistry(t)
	ctx := context.Background()
	data := bytes.Repeat([]byte{0xa5}, fileChunkBytes+17)
	data[0], data[len(data)-1] = 0, 42
	upload := func(source io.Reader, size int64, guard func() error) (FileMetadata, error) {
		return r.UploadFileAsOperator(ctx, "alpha", version, "object", "operator", "file.bin", "application/octet-stream", source, size, guard)
	}
	meta, err := upload(bytes.NewReader(data), int64(len(data)), nil)
	if err != nil || meta.SizeBytes != int64(len(data)) || meta.Encoding != "binary" || meta.Filename != "file.bin" || meta.CreatedAt == "" {
		t.Fatalf("upload: %+v %v", meta, err)
	}
	check := func(want []byte) {
		t.Helper()
		var received bytes.Buffer
		if err := r.DownloadFileAsOperator(ctx, "alpha", version, "object", "operator", func(m FileMetadata) io.Writer {
			if m.SizeBytes != int64(len(want)) {
				t.Fatalf("wrong size: %+v", m)
			}
			return &received
		}); err != nil || !bytes.Equal(received.Bytes(), want) {
			t.Fatalf("download mismatch: %v", err)
		}
	}
	check(data)
	if err := r.DownloadFileAsOperator(ctx, "beta", version, "object", "operator", func(FileMetadata) io.Writer {
		t.Fatal("cross-project download started")
		return io.Discard
	}); !errors.Is(err, ErrStoredResourceMissing) {
		t.Fatalf("foreign object: %v", err)
	}
	failure := errors.New("operator session revoked")
	for _, failAt := range []int{1, 2} {
		calls := 0
		_, err := upload(strings.NewReader("replacement"), 11, func() error {
			calls++
			if calls == failAt {
				return failure
			}
			return nil
		})
		if !errors.Is(err, failure) || calls != failAt {
			t.Fatalf("guard %d: %v, calls=%d", failAt, err, calls)
		}
		check(data)
	}
	for _, tc := range []struct {
		writer io.Writer
		want   error
	}{
		{binaryTestWriter(func([]byte) (int, error) { return 0, failure }), failure},
		{binaryTestWriter(func(p []byte) (int, error) { return len(p) - 1, nil }), io.ErrShortWrite},
	} {
		if err := r.DownloadFileAsOperator(ctx, "alpha", version, "object", "operator", func(FileMetadata) io.Writer { return tc.writer }); !errors.Is(err, tc.want) {
			t.Fatalf("writer failure: %v", err)
		}
	}
	if _, err := s.db.Exec("DELETE FROM plugin_file_chunks WHERE project_id='alpha' AND file_key='object' AND chunk_index=1"); err != nil {
		t.Fatal(err)
	}
	if err := r.DownloadFileAsOperator(ctx, "alpha", version, "object", "operator", func(FileMetadata) io.Writer { return io.Discard }); !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("truncated object: %v", err)
	}
	if err := s.PutFile("alpha", "object", "legacy", false); err != nil {
		t.Fatal(err)
	}
	check([]byte("legacy"))
	var count int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM plugin_file_chunks WHERE project_id='alpha' AND file_key='object'").Scan(&count); err != nil || count != 0 {
		t.Fatalf("legacy overwrite left chunks: %d %v", count, err)
	}
}

func TestBinaryFileAdmissionAndStreamFailures(t *testing.T) {
	r, _, version := binaryTestRegistry(t)
	ctx := context.Background()
	failure := errors.New("reader failed")
	upload := func(ctx context.Context, project, key, actor string, source io.Reader, size int64) error {
		_, err := r.UploadFileAsOperator(ctx, project, version, key, actor, "file.bin", "", source, size, nil)
		return err
	}
	for _, key := range []string{"", " bad", strings.Repeat("a", 1025)} {
		if err := upload(ctx, "alpha", key, "operator", strings.NewReader("a"), 1); !errors.Is(err, ErrInvalidWorkerRequest) {
			t.Fatalf("bad key: %v", err)
		}
	}
	if err := upload(ctx, "alpha", "object", "", strings.NewReader("a"), 1); !errors.Is(err, ErrInvalidWorkerRequest) {
		t.Fatalf("bad actor: %v", err)
	}
	if err := upload(ctx, "alpha", "object", "operator", strings.NewReader("a"), 2); !errors.Is(err, ErrInvalidWorkerRequest) {
		t.Fatalf("size mismatch: %v", err)
	}
	if err := upload(ctx, "alpha", "object", "operator", iotest.ErrReader(failure), -1); !errors.Is(err, failure) {
		t.Fatalf("reader failure: %v", err)
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if err := upload(cancelled, "alpha", "object", "operator", strings.NewReader("a"), 1); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled stream: %v", err)
	}
	if err := r.Configure("alpha", "files", version, map[string]any{"maxUploadBytes": 3}); err != nil {
		t.Fatal(err)
	}
	for _, size := range []int64{4, -1} {
		if err := upload(ctx, "alpha", "object", "operator", strings.NewReader("abcd"), size); !errors.Is(err, ErrFileTooLarge) {
			t.Fatalf("limit: %v", err)
		}
	}
	for i := 0; i < cap(fileUploadSlots); i++ {
		fileUploadSlots <- struct{}{}
	}
	err := upload(ctx, "alpha", "object", "operator", strings.NewReader("a"), 1)
	for i := 0; i < cap(fileUploadSlots); i++ {
		<-fileUploadSlots
	}
	if !errors.Is(err, ErrFileTransferBusy) {
		t.Fatalf("capacity: %v", err)
	}
	r.UnbindWorker("alpha", "files", version)
	if err := upload(ctx, "alpha", "object", "operator", strings.NewReader("a"), 1); !errors.Is(err, ErrWorkerNotReady) {
		t.Fatalf("worker readiness: %v", err)
	}
	if err := r.DownloadFileAsOperator(ctx, "alpha", version, "object", "operator", func(FileMetadata) io.Writer { return io.Discard }); !errors.Is(err, ErrWorkerNotReady) {
		t.Fatalf("download readiness: %v", err)
	}
}
