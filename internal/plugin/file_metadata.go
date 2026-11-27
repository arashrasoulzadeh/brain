package plugin

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"mime"
	"net/http"
	"path"
	"strings"
	"time"
	"unicode"
)

// FileMetadata is host-derived except for the original filename and declared
// content type, which are untrusted descriptive hints. Downloads are attachments.
type FileMetadata struct {
	Filename            string `json:"filename"`
	Extension           string `json:"extension"`
	ContentType         string `json:"content_type"`
	DeclaredContentType string `json:"declared_content_type,omitempty"`
	SizeBytes           int64  `json:"size_bytes"`
	StoredBytes         int64  `json:"stored_bytes"`
	SHA256              string `json:"sha256"`
	Encoding            string `json:"encoding"`
	CreatedAt           string `json:"created_at,omitempty"`
	UpdatedAt           string `json:"updated_at,omitempty"`
	PublishedAt         string `json:"published_at,omitempty"`
}

func fileBytes(data string) ([]byte, string) {
	if encoded, found := strings.CutPrefix(data, adminFileDataPrefix); found {
		if decoded, err := base64.StdEncoding.Strict().DecodeString(encoded); err == nil && base64.StdEncoding.EncodeToString(decoded) == encoded {
			return decoded, "base64"
		}
	}
	return []byte(data), "utf8"
}

func safeFilename(value string) string {
	name := path.Base(strings.ReplaceAll(value, "\\", "/"))
	name = strings.TrimSpace(strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, name))
	if name == "" || name == "." || name == ".." || name == "/" || len(name) > 255 {
		return ""
	}
	return name
}

func deriveFileMetadata(key, data, filename, declared string) FileMetadata {
	if filename = safeFilename(filename); filename == "" {
		filename = safeFilename(key)
	}
	if filename == "" {
		filename = "download.bin"
	}
	extension := strings.TrimPrefix(strings.ToLower(path.Ext(filename)), ".")
	if path.Ext(filename) == filename {
		extension = ""
	}
	bytes, encoding := fileBytes(data)
	detected := http.DetectContentType(bytes)
	sum := sha256.Sum256(bytes)
	declaredType := ""
	if len(declared) <= 255 {
		if parsed, _, err := mime.ParseMediaType(declared); err == nil && strings.Contains(parsed, "/") {
			declaredType = parsed
		}
	}
	return FileMetadata{Filename: filename, Extension: extension, ContentType: detected, DeclaredContentType: declaredType, SizeBytes: int64(len(bytes)), StoredBytes: int64(len(data)), SHA256: hex.EncodeToString(sum[:]), Encoding: encoding}
}

func (s *SQLiteStorage) initializeFileMetadata() error {
	// Migrate old SQLite installations in place; no data is rewritten.
	rows, err := s.db.Query("PRAGMA table_info(plugin_file_rows)")
	if err != nil {
		return err
	}
	found := false
	for rows.Next() {
		var cid, notNull, primary int
		var name, kind string
		var defaultValue any
		if err := rows.Scan(&cid, &name, &kind, &notNull, &defaultValue, &primary); err != nil {
			rows.Close()
			return err
		}
		found = found || name == "metadata_json"
	}
	err = rows.Err()
	rows.Close()
	if err != nil || found {
		return err
	}
	_, err = s.db.Exec("ALTER TABLE plugin_file_rows ADD COLUMN metadata_json TEXT NOT NULL DEFAULT ''")
	return err
}

func (s *SQLiteStorage) fileWithMetadata(projectID, key string) (string, bool, FileMetadata, error) {
	var data, encoded string
	var published bool
	err := s.db.QueryRow("SELECT data,published,metadata_json FROM plugin_file_rows WHERE project_id=? AND file_key=?", projectID, key).Scan(&data, &published, &encoded)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, FileMetadata{}, nil
	}
	if err != nil {
		return "", false, FileMetadata{}, err
	}
	meta := deriveFileMetadata(key, data, "", "")
	if encoded != "" {
		if err := json.Unmarshal([]byte(encoded), &meta); err != nil {
			return "", false, FileMetadata{}, err
		}
	}
	return data, published, meta, nil
}

func putSQLiteFileMetadata(db sqliteExecutor, projectID, key, data string, published bool, meta FileMetadata) error {
	encoded, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	_, err = db.Exec(`INSERT INTO plugin_file_rows(project_id,file_key,data,published,metadata_json) VALUES(?,?,?,?,?) ON CONFLICT(project_id,file_key) DO UPDATE SET data=excluded.data,published=excluded.published,metadata_json=excluded.metadata_json`, projectID, key, data, published, string(encoded))
	return err
}

func fileTimestamp() string {
	return time.Now().UTC().Format(time.RFC3339Nano)
}
