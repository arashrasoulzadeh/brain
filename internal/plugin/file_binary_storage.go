package plugin

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
)

var ErrFileTooLarge = errors.New("file exceeds project upload limit")
var ErrFileTransferBusy = errors.New("file transfer capacity reached")

const fileChunkBytes = 1 << 20

// Bound staging disk use across concurrent authenticated uploads.
var fileUploadSlots = make(chan struct{}, 2)

func (s *SQLiteStorage) initializeFileChunks() error {
	_, err := s.db.Exec(`CREATE TABLE IF NOT EXISTS plugin_file_chunks (
		project_id TEXT NOT NULL, file_key TEXT NOT NULL, chunk_index INTEGER NOT NULL,
		data BLOB NOT NULL, PRIMARY KEY(project_id,file_key,chunk_index));
	CREATE TRIGGER IF NOT EXISTS plugin_file_chunks_legacy_overwrite AFTER UPDATE OF data ON plugin_file_rows
	WHEN NEW.data <> '' BEGIN
		DELETE FROM plugin_file_chunks WHERE project_id=NEW.project_id AND file_key=NEW.file_key;
	END;
	CREATE TRIGGER IF NOT EXISTS plugin_file_chunks_delete AFTER DELETE ON plugin_file_rows BEGIN
		DELETE FROM plugin_file_chunks WHERE project_id=OLD.project_id AND file_key=OLD.file_key;
	END;`)
	return err
}

func (r *Registry) fileTransferRequest(projectID, version, key, operation, actorID string) (WorkerRequest, error) {
	if !validAccessIdentity(actorID) || strings.TrimSpace(key) != key || key == "" || len(key) > 1024 {
		return WorkerRequest{}, ErrInvalidWorkerRequest
	}
	payload, _ := json.Marshal(map[string]string{"key": key})
	request := WorkerRequest{Protocol: r.HostAPI, ProjectID: projectID, PluginID: "files", Version: version, Operation: operation, Payload: payload, accessPrincipal: "operator:" + actorID}
	return request, ValidateWorkerRequest(request, r.HostAPI)
}

// Caller holds readiness and installation read locks. Use the same exact
// resource, declared capability and project policy as the JSON file operation.
func (r *Registry) fileTransferAdmissionLocked(request WorkerRequest) (*SQLiteStorage, int64, error) {
	if err := r.authorizeInvocationLocked(request, false); err != nil {
		return nil, 0, err
	}
	manifest := r.registered["files@"+request.Version].Manifest
	route, exists := manifest.Route(request.Operation)
	installation := r.installations[request.ProjectID+":files@"+request.Version]
	if !exists || route.ResourceField != "key" || route.Capability != request.Operation || !manifest.AdminOperation("operations", request.Operation) || installation.Draining {
		return nil, 0, ErrGrantDenied
	}
	r.workerMu.RLock()
	worker := r.workers[request.ProjectID+":files@"+request.Version]
	ready := worker != nil && worker.Ready()
	r.workerMu.RUnlock()
	if !ready {
		return nil, 0, ErrWorkerNotReady
	}
	storage, ok := r.storage.(*SQLiteStorage)
	if !ok || storage == nil {
		return nil, 0, ErrLifecycleStorage
	}
	limit := defaultFileUploadLimit
	if configured, ok := configuredFileUploadLimit(installation.Config); ok {
		limit = configured
	}
	return storage, int64(limit), nil
}

func (r *Registry) fileTransferAdmission(request WorkerRequest) (*SQLiteStorage, int64, error) {
	r.readinessMu.RLock()
	defer r.readinessMu.RUnlock()
	r.installationMu.RLock()
	defer r.installationMu.RUnlock()
	return r.fileTransferAdmissionLocked(request)
}

// UploadFileAsOperator stages a bounded stream outside lifecycle locks, then
// atomically commits file metadata and fixed-size chunks after reauthorization.
// No complete file is held in Go memory or sent through a worker JSON message.
func (r *Registry) UploadFileAsOperator(ctx context.Context, projectID, version, key, actorID, filename, contentType string, source io.Reader, declaredSize int64, guard func() error) (FileMetadata, error) {
	request, err := r.fileTransferRequest(projectID, version, key, "files.upload", actorID)
	if err != nil {
		return FileMetadata{}, err
	}
	storage, limit, err := r.fileTransferAdmission(request)
	if err != nil {
		return FileMetadata{}, err
	}
	if declaredSize > limit {
		return FileMetadata{}, ErrFileTooLarge
	}
	select {
	case fileUploadSlots <- struct{}{}:
		defer func() { <-fileUploadSlots }()
	default:
		return FileMetadata{}, ErrFileTransferBusy
	}
	temp, err := os.CreateTemp("", "brain-upload-*")
	if err != nil {
		return FileMetadata{}, err
	}
	defer os.Remove(temp.Name())
	defer temp.Close()
	hash := sha256.New()
	size, err := io.Copy(io.MultiWriter(temp, hash), io.LimitReader(source, limit+1))
	if err != nil {
		return FileMetadata{}, err
	}
	if size > limit {
		return FileMetadata{}, ErrFileTooLarge
	}
	if declaredSize >= 0 && size != declaredSize {
		return FileMetadata{}, ErrInvalidWorkerRequest
	}
	if err := ctx.Err(); err != nil {
		return FileMetadata{}, err
	}
	sniff := make([]byte, 512)
	n, err := temp.ReadAt(sniff, 0)
	if err != nil && err != io.EOF {
		return FileMetadata{}, err
	}
	meta := deriveFileMetadata(key, "", filename, contentType)
	meta.ContentType, meta.Encoding = http.DetectContentType(sniff[:n]), "binary"
	meta.SizeBytes, meta.StoredBytes, meta.SHA256 = size, size, hex.EncodeToString(hash.Sum(nil))
	if _, err := temp.Seek(0, io.SeekStart); err != nil {
		return FileMetadata{}, err
	}

	_, unlock := r.lockStorageInvocation(request)
	defer unlock()
	r.readinessMu.RLock()
	defer r.readinessMu.RUnlock()
	r.installationMu.RLock()
	defer r.installationMu.RUnlock()
	current, currentLimit, err := r.fileTransferAdmissionLocked(request)
	if err != nil {
		return FileMetadata{}, err
	}
	if current != storage {
		return FileMetadata{}, ErrLifecycleStorage
	}
	if size > currentLimit {
		return FileMetadata{}, ErrFileTooLarge
	}
	if guard != nil {
		if err := guard(); err != nil {
			return FileMetadata{}, err
		}
	}
	_, _, previous, err := storage.fileWithMetadata(projectID, key)
	if err != nil {
		return FileMetadata{}, err
	}
	meta.CreatedAt, meta.UpdatedAt = previous.CreatedAt, fileTimestamp()
	if meta.CreatedAt == "" {
		meta.CreatedAt = meta.UpdatedAt
	}
	tx, err := storage.db.BeginTx(ctx, nil)
	if err != nil {
		return FileMetadata{}, err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, "DELETE FROM plugin_file_chunks WHERE project_id=? AND file_key=?", projectID, key); err != nil {
		return FileMetadata{}, err
	}
	buffer := make([]byte, fileChunkBytes)
	for index := 0; ; index++ {
		count, readErr := io.ReadFull(temp, buffer)
		if count > 0 {
			if _, err := tx.ExecContext(ctx, "INSERT INTO plugin_file_chunks(project_id,file_key,chunk_index,data) VALUES(?,?,?,?)", projectID, key, index, buffer[:count]); err != nil {
				return FileMetadata{}, err
			}
		}
		if readErr == io.EOF || readErr == io.ErrUnexpectedEOF {
			break
		}
		if readErr != nil {
			return FileMetadata{}, readErr
		}
	}
	if err := putSQLiteFileMetadata(tx, projectID, key, "", false, meta); err != nil {
		return FileMetadata{}, err
	}
	if guard != nil {
		if err := guard(); err != nil {
			return FileMetadata{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return FileMetadata{}, err
	}
	return meta, nil
}

// DownloadFileAsOperator opens one consistent file snapshot. The caller must
// send headers only through begin, after authorization and existence checks.
func (r *Registry) DownloadFileAsOperator(ctx context.Context, projectID, version, key, actorID string, begin func(FileMetadata) io.Writer) error {
	request, err := r.fileTransferRequest(projectID, version, key, "files.download", actorID)
	if err != nil {
		return err
	}
	_, unlock := r.lockStorageInvocation(request)
	defer unlock()
	storage, _, err := r.fileTransferAdmission(request)
	if err != nil {
		return err
	}
	data, _, meta, err := storage.fileWithMetadata(projectID, key)
	if err != nil {
		return err
	}
	if data == "" && meta.Encoding != "binary" {
		return ErrStoredResourceMissing
	}
	if meta.Encoding != "binary" {
		bytes, _ := fileBytes(data)
		_, err = begin(meta).Write(bytes)
		return err
	}
	rows, err := storage.db.QueryContext(ctx, "SELECT data FROM plugin_file_chunks WHERE project_id=? AND file_key=? ORDER BY chunk_index", projectID, key)
	if err != nil {
		return err
	}
	defer rows.Close()
	writer := begin(meta)
	var total int64
	for rows.Next() {
		var chunk []byte
		if err := rows.Scan(&chunk); err != nil {
			return err
		}
		n, err := writer.Write(chunk)
		total += int64(n)
		if err != nil {
			return err
		}
		if n != len(chunk) {
			return io.ErrShortWrite
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if total != meta.SizeBytes {
		return io.ErrUnexpectedEOF
	}
	return nil
}
