package plugin

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

const authAccountsNamespace = "auth.accounts"
const authSessionsNamespace = "auth.sessions"

var errAuthCredentials = errors.New("invalid credentials")
var errAuthSession = errors.New("invalid session")

type durableAuthAccount struct {
	Username     string `json:"username"`
	UserID       string `json:"userId"`
	PasswordHash string `json:"passwordHash"`
}

type durableAuthSession struct {
	ProjectID string `json:"projectId"`
	Username  string `json:"username"`
	ExpiresAt int64  `json:"expiresAt"`
}

type authAccountSummary struct {
	Username       string `json:"username"`
	UserID         string `json:"userId"`
	Guest          bool   `json:"guest"`
	ActiveSessions int    `json:"activeSessions"`
}

type authAccountPage struct {
	Rows []authAccountSummary `json:"rows"`
	Next string               `json:"next"`
}

func authTokenDigest(token string) string {
	digest := sha256.Sum256([]byte(token))
	return hex.EncodeToString(digest[:])
}

func authUsername(value string) bool {
	return len(value) <= 128 && isValidProjectID(value)
}

func readAuthState[T any](s *SQLiteStorage, project, namespace, key string) (*T, error) {
	value, err := s.State(project, namespace, key)
	if err != nil || value == nil {
		return nil, err
	}
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var record T
	if err := json.Unmarshal(data, &record); err != nil {
		return nil, ErrModelSchema
	}
	return &record, nil
}

func (s *SQLiteStorage) authAccountPage(project, after string, now int64) (authAccountPage, error) {
	result := authAccountPage{Rows: []authAccountSummary{}}
	// Only the redacted projection is selected. Neither password hashes nor
	// session digest keys are ever sent to the administrative worker.
	rows, err := s.db.Query(`WITH active AS (
		SELECT json_extract(value_json,'$.username') AS username, COUNT(*) AS n
		FROM plugin_state WHERE project_id=? AND namespace='auth.sessions' AND json_extract(value_json,'$.expiresAt')>?
		GROUP BY username
	), accounts AS (
		SELECT state_key AS username, json_extract(value_json,'$.userId') AS user_id
		FROM plugin_state WHERE project_id=? AND namespace='auth.accounts'
	), users AS (
		SELECT accounts.username,accounts.user_id,0 AS guest,COALESCE(active.n,0) AS sessions
		FROM accounts LEFT JOIN active ON accounts.username=active.username
		UNION ALL
		SELECT active.username,? || ':' || active.username,1,active.n
		FROM active LEFT JOIN accounts ON active.username=accounts.username
		WHERE accounts.username IS NULL AND substr(active.username,1,6)='guest-'
	)
	SELECT username,user_id,guest,sessions FROM users WHERE username>? ORDER BY username LIMIT 51`, project, now, project, project, after)
	if err != nil {
		return result, err
	}
	defer rows.Close()
	for rows.Next() {
		var row authAccountSummary
		if err := rows.Scan(&row.Username, &row.UserID, &row.Guest, &row.ActiveSessions); err != nil {
			return authAccountPage{}, err
		}
		if len(result.Rows) == 50 {
			result.Next = result.Rows[49].Username
			break
		}
		result.Rows = append(result.Rows, row)
	}
	return result, rows.Err()
}

// Only the account or session needed for this operation is supplied. The
// project worker cannot read another project's credential store.
func (p legacyStoragePlanner) prepareAuthStorage(request WorkerRequest) (WorkerRequest, func(*WorkerResponse, []observationDelivery) error, error) {
	if p.storage == nil {
		return request, nil, ErrLifecycleStorage
	}
	var input map[string]any
	if json.Unmarshal(request.Payload, &input) != nil || input == nil {
		return request, nil, ErrInvalidWorkerRequest
	}
	project := request.ProjectID
	payload := map[string]any{}
	private := map[string]any{}
	var username, token string
	var account *durableAuthAccount
	var session *durableAuthSession
	var page authAccountPage
	var err error
	switch request.Operation {
	case "auth.signup", "auth.signin":
		username, _ = input["username"].(string)
		password, _ := input["password"].(string)
		if request.Operation == "auth.signin" {
			if input["username"] == nil {
				username, _ = input["provider"].(string)
			}
			if input["password"] == nil {
				password, _ = input["credential"].(string)
			}
		}
		if !authUsername(username) || strings.HasPrefix(username, "guest-") || len(password) > 1024 {
			return request, nil, errAuthCredentials
		}
		payload["username"], payload["password"] = username, password
		account, err = readAuthState[durableAuthAccount](p.storage, project, authAccountsNamespace, username)
		if err != nil {
			return request, nil, err
		}
		if request.Operation == "auth.signup" && account != nil {
			return request, nil, errAuthCredentials
		}
		if request.Operation == "auth.signin" {
			if account == nil || account.Username != username || account.UserID != project+":"+username || account.PasswordHash == "" {
				return request, nil, errAuthCredentials
			}
			private["account"] = account
		}
	case "auth.session", "auth.identity":
		token, _ = input["token"].(string)
		if token == "" || len(token) > 1024 {
			return request, nil, errAuthSession
		}
		session, err = readAuthState[durableAuthSession](p.storage, project, authSessionsNamespace, authTokenDigest(token))
		if err != nil {
			return request, nil, err
		}
		if session == nil || session.ProjectID != project || !authUsername(session.Username) || session.ExpiresAt <= time.Now().UnixMilli() {
			return request, nil, errAuthSession
		}
		username = session.Username
		payload["token"], private["session"] = token, session
		if request.Operation == "auth.identity" && !strings.HasPrefix(username, "guest-") {
			account, err = readAuthState[durableAuthAccount](p.storage, project, authAccountsNamespace, username)
			if err != nil {
				return request, nil, err
			}
			if account == nil || account.Username != username || account.UserID != project+":"+username {
				return request, nil, errAuthSession
			}
			// Identity lookup needs only the public identity, not a hash.
			private["account"] = durableAuthAccount{Username: account.Username, UserID: account.UserID}
		}
	case "auth.guest":
	case "auth.admin.accounts":
		after, ok := input["after"].(string)
		if (input["after"] != nil && !ok) || len(after) > 1000 {
			return request, nil, ErrInvalidWorkerRequest
		}
		page, err = p.storage.authAccountPage(project, after, time.Now().UnixMilli())
		if err != nil {
			return request, nil, err
		}
		payload["after"], private["page"] = after, page
	case "auth.admin.sessions.revoke":
		username, _ = input["username"].(string)
		if !authUsername(username) {
			return request, nil, ErrInvalidWorkerRequest
		}
		payload["username"] = username
	default:
		return request, nil, ErrInvalidWorkerRequest
	}
	payload["_brainAuth"] = private
	prepared, err := json.Marshal(payload)
	if err != nil {
		return request, nil, err
	}
	request.Payload = prepared
	return request, func(response *WorkerResponse, deliveries []observationDelivery) error {
		var public any
		var mutation func(sqliteExecutor) error
		switch request.Operation {
		case "auth.signup":
			var result struct {
				Result       durableAuthAccount `json:"result"`
				PasswordHash string             `json:"passwordHash"`
			}
			if json.Unmarshal(response.Payload, &result) != nil || result.Result.Username != username || result.Result.UserID != project+":"+username ||
				!strings.HasPrefix(result.PasswordHash, "$argon2id$") || len(result.PasswordHash) > 1024 {
				return ErrInvalidWorkerRequest
			}
			data, err := json.Marshal(durableAuthAccount{Username: username, UserID: project + ":" + username, PasswordHash: result.PasswordHash})
			if err != nil {
				return err
			}
			mutation = func(db sqliteExecutor) error {
				return compareAndSwapPluginState(db, project, authAccountsNamespace, username, "", false, string(data))
			}
			public = map[string]string{"username": username, "userId": project + ":" + username}
		case "auth.signin", "auth.guest":
			var result struct {
				Result struct {
					Token    string `json:"token"`
					Username string `json:"username"`
					UserID   string `json:"userId"`
					Guest    bool   `json:"guest"`
				} `json:"result"`
				ExpiresAt int64 `json:"expiresAt"`
			}
			if json.Unmarshal(response.Payload, &result) != nil {
				return ErrInvalidWorkerRequest
			}
			now := time.Now().UnixMilli()
			if result.ExpiresAt <= now || result.ExpiresAt > now+int64(24*time.Hour/time.Millisecond) ||
				len(result.Result.Token) != 64 || !authUsername(result.Result.Username) {
				return ErrInvalidWorkerRequest
			}
			if _, err := hex.DecodeString(result.Result.Token); err != nil {
				return ErrInvalidWorkerRequest
			}
			if request.Operation == "auth.guest" {
				username = result.Result.Username
				if !strings.HasPrefix(username, "guest-") || !result.Result.Guest || result.Result.UserID != project+":"+username {
					return ErrInvalidWorkerRequest
				}
				public = map[string]any{"token": result.Result.Token, "username": username, "userId": project + ":" + username, "guest": true}
			} else {
				if result.Result.Username != username || result.Result.Guest {
					return ErrInvalidWorkerRequest
				}
				public = map[string]string{"token": result.Result.Token, "username": username}
			}
			data, err := json.Marshal(durableAuthSession{ProjectID: project, Username: username, ExpiresAt: result.ExpiresAt})
			if err != nil {
				return err
			}
			digest := authTokenDigest(result.Result.Token)
			mutation = func(db sqliteExecutor) error {
				// Opportunistically remove expired session records.
				if _, err := db.Exec(`DELETE FROM plugin_state WHERE project_id=? AND namespace=? AND json_extract(value_json,'$.expiresAt')<=?`, project, authSessionsNamespace, now); err != nil {
					return err
				}
				return compareAndSwapPluginState(db, project, authSessionsNamespace, digest, "", false, string(data))
			}
		case "auth.session", "auth.identity":
			// Recheck the stored credential before exposing the verified identity.
			current, err := readAuthState[durableAuthSession](p.storage, project, authSessionsNamespace, authTokenDigest(token))
			if err != nil {
				return err
			}
			if current == nil || *current != *session || current.ExpiresAt <= time.Now().UnixMilli() {
				return errAuthSession
			}
			if request.Operation == "auth.identity" {
				public = map[string]string{"userId": project + ":" + username}
			} else {
				public = map[string]string{"username": username}
			}
		case "auth.admin.accounts":
			public = page
		case "auth.admin.sessions.revoke":
			mutation = func(db sqliteExecutor) error {
				deleted, err := db.Exec(`DELETE FROM plugin_state WHERE project_id=? AND namespace=? AND json_extract(value_json,'$.username')=?`, project, authSessionsNamespace, username)
				if err != nil {
					return err
				}
				count, err := deleted.RowsAffected()
				if err != nil {
					return err
				}
				public = map[string]any{"username": username, "revoked": count}
				return nil
			}
		}
		if mutation != nil || len(deliveries) > 0 {
			tx, err := p.storage.db.Begin()
			if err != nil {
				return err
			}
			defer tx.Rollback()
			if mutation != nil {
				if err := mutation(tx); err != nil {
					return err
				}
			}
			for _, delivery := range deliveries {
				if err := enqueueSQLiteObservation(tx, delivery, time.Now()); err != nil {
					return err
				}
			}
			if err := tx.Commit(); err != nil {
				return err
			}
		}
		// Do not return private credential material, even to administrators.
		encoded, err := json.Marshal(public)
		if err != nil {
			return err
		}
		response.Payload = encoded
		return nil
	}, nil
}
