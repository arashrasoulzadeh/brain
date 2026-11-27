package app

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strconv"
	"strings"

	"brain/internal/admin"
	"brain/internal/operator"
	"brain/internal/plugin"
)

func fileTransferError(w http.ResponseWriter, err error) {
	status := http.StatusForbidden
	switch {
	case errors.Is(err, plugin.ErrFileTooLarge):
		status = http.StatusRequestEntityTooLarge
	case errors.Is(err, plugin.ErrFileTransferBusy):
		status = http.StatusTooManyRequests
	case errors.Is(err, plugin.ErrInvalidWorkerRequest):
		status = http.StatusBadRequest
	case errors.Is(err, plugin.ErrStoredResourceMissing):
		status = http.StatusNotFound
	}
	// The binary GET bypasses the response buffer, so provide its own safe
	// error envelope. Do not expose storage paths or database error details.
	message := http.StatusText(status)
	if status == http.StatusRequestEntityTooLarge {
		message = "File exceeds project upload limit"
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(APIMessage{Status: "error", Code: http.StatusText(status), Message: message})
}

func AuthenticatedFileTransferHandler(service *admin.LifecycleService, sessions *operator.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actor, ok := operatorActorFromRequest(sessions, r)
		if !ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(APIMessage{Status: "error", Message: "Operator session required"})
			return
		}
		p, ok := projectIdentity(r)
		if !ok {
			http.NotFound(w, r)
			return
		}
		query := r.URL.Query()
		for _, values := range query {
			if len(values) != 1 {
				fileTransferError(w, plugin.ErrInvalidWorkerRequest)
				return
			}
		}
		version, key := query.Get("version"), query.Get("key")
		if version == "" || key == "" {
			fileTransferError(w, plugin.ErrInvalidWorkerRequest)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		if r.Method == http.MethodPost {
			if query.Get("confirmed") != "true" || len(query.Get("filename")) > 1024 || len(query.Get("contentType")) > 255 {
				fileTransferError(w, plugin.ErrInvalidWorkerRequest)
				return
			}
			kind, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
			if err != nil || !strings.EqualFold(kind, "application/octet-stream") {
				http.Error(w, "raw binary upload requires application/octet-stream", http.StatusUnsupportedMediaType)
				return
			}
			meta, err := service.UploadFile(actor, r.Context(), p.ID, version, key, query.Get("filename"), query.Get("contentType"), r.Body, r.ContentLength)
			if err != nil {
				fileTransferError(w, err)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"key": key, "uploaded": true, "published": false, "metadata": meta})
			return
		}
		started := false
		err := service.DownloadFile(actor, r.Context(), p.ID, version, key, func(meta plugin.FileMetadata) io.Writer {
			started = true
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": meta.Filename}))
			w.Header().Set("Content-Length", strconv.FormatInt(meta.SizeBytes, 10))
			w.WriteHeader(http.StatusOK)
			return w
		})
		if err != nil && !started {
			fileTransferError(w, err)
		}
	})
}
