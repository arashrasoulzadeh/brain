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
	Suspended    bool   `json:"suspended,omitempty"`
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
	Suspended      bool   `json:"suspended,omitempty"`
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
	return s.authAccountSearch(project, after, "", "", now, 50)
}

func (s *SQLiteStorage) authAccountSearch(project, after, search, kind string, now int64, limit int) (authAccountPage, error) {
	result := authAccountPage{Rows: []authAccountSummary{}}
	// Only the redacted projection is selected. Neither password hashes nor
	// session digest keys are ever sent to the administrative worker.
	rows, err := s.db.Query(`WITH active AS (
		SELECT json_extract(value_json,'$.username') AS username, COUNT(*) AS n
		FROM plugin_state WHERE project_id=? AND namespace='auth.sessions' AND json_extract(value_json,'$.expiresAt')>?
		GROUP BY username
	), accounts AS (
		SELECT state_key AS username, json_extract(value_json,'$.userId') AS user_id, COALESCE(json_extract(value_json,'$.suspended'),0) AS suspended
		FROM plugin_state WHERE project_id=? AND namespace='auth.accounts'
	), users AS (
		SELECT accounts.username,accounts.user_id,0 AS guest,COALESCE(active.n,0) AS sessions,accounts.suspended
		FROM accounts LEFT JOIN active ON accounts.username=active.username
		UNION ALL
		SELECT active.username,? || ':' || active.username,1,active.n,0
		FROM active LEFT JOIN accounts ON active.username=accounts.username
		WHERE accounts.username IS NULL AND substr(active.username,1,6)='guest-'
	)
	SELECT username,user_id,guest,sessions,suspended FROM users WHERE username>?
		AND (instr(lower(username),lower(?))>0 OR instr(lower(user_id),lower(?))>0)
		AND (?=\x27\x27 OR (?=\x27registered\x27 AND guest=0) OR (?=\x27guest\x27 AND guest=1) OR (?=\x27suspended\x27 AND suspended=1))
		ORDER BY username LIMIT ?`, project, now, project, project, after, search, search, kind, kind, kind, kind, limit+1)
	if err != nil {
		return result, err
	}
	defer rows.Close()
	for rows.Next() {
		var row authAccountSummary
		if err := rows.Scan(&row.Username, &row.UserID, &row.Guest, &row.ActiveSessions, &row.Suspended); err != nil {
			return authAccountPage{}, err
		}
		if len(result.Rows) == limit {
			result.Next = result.Rows[limit-1].Username
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
	var sessionSummary authSessionSummary
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
			if account == nil || account.Username != username || account.UserID != project+":"+username || account.PasswordHash == "" || account.Suspended {
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
		if !strings.HasPrefix(username, "guest-") {
			account, err = readAuthState[durableAuthAccount](p.storage, project, authAccountsNamespace, username)
			if err != nil {
				return request, nil, err
			}
			if account == nil || account.Username != username || account.UserID != project+":"+username || account.Suspended {
				return request, nil, errAuthSession
			}
			// Identity lookup needs only the public identity, not a hash.
			if request.Operation == "auth.identity" {
				private["account"] = durableAuthAccount{Username: account.Username, UserID: account.UserID}
			}
		}
	case "auth.guest":
	case "auth.admin.accounts":
		after, ok := input["after"].(string)
		if (input["after"] != nil && !ok) || len(after) > 1000 {
			return request, nil, ErrInvalidWorkerRequest
		}
		search, searchOK := input["search"].(string)
		kind, kindOK := input["kind"].(string)
		if (input["search"] != nil && !searchOK) || len(search) > 256 || strings.ContainsRune(search, 0) || (input["kind"] != nil && !kindOK) || (kind != "" && kind != "registered" && kind != "guest" && kind != "suspended") {
			return request, nil, ErrInvalidWorkerRequest
		}
		limit, limitErr := boundedStorageInt(input, "limit", 50, 1, 50)
		if limitErr != nil {
			return request, nil, limitErr
		}
		page, err = p.storage.authAccountSearch(project, after, strings.TrimSpace(search), kind, time.Now().UnixMilli(), limit)
		if err != nil {
			return request, nil, err
		}
		payload["after"], private["page"] = after, page
	case "auth.admin.sessions.inspect":
		username, _ = input["username"].(string)
		if !authUsername(username) {
			return request, nil, ErrInvalidWorkerRequest
		}
		sessionSummary, err = p.storage.authSessionSummary(project, username, time.Now().UnixMilli())
		if err != nil {
			return request, nil, err
		}
		payload["username"], private["sessions"] = username, sessionSummary
	case "auth.admin.accounts.suspend", "auth.admin.accounts.reactivate":
		username, _ = input["username"].(string)
		if !authUsername(username) || strings.HasPrefix(username, "guest-") {
			return request, nil, ErrInvalidWorkerRequest
		}
		account, err = readAuthState[durableAuthAccount](p.storage, project, authAccountsNamespace, username)
		if err != nil {
			return request, nil, err
		}
		if account == nil || account.Username != username || account.UserID != project+":"+username {
			return request, nil, ErrStoredResourceMissing
		}
		payload["username"] = username
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
				if request.Operation == "auth.signin" {
					var suspended bool
					if err := db.QueryRow(`SELECT COALESCE(json_extract(value_json,\x27$.suspended\x27),0) FROM plugin_state WHERE project_id=? AND namespace=? AND state_key=?`, project, authAccountsNamespace, username).Scan(&suspended); err != nil || suspended {
						return errAuthCredentials
					}
				}
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
			var identity map[string]string
			if json.Unmarshal(response.Payload, &identity) != nil || len(identity) != 1 {
				return ErrInvalidWorkerRequest
			}
			if request.Operation == "auth.identity" {
				if identity["userId"] != project+":"+username {
					return ErrInvalidWorkerRequest
				}
				public = map[string]string{"userId": project + ":" + username}
			} else {
				if identity["username"] != username {
					return ErrInvalidWorkerRequest
				}
				public = map[string]string{"username": username}
			}
		case "auth.admin.accounts":
			public = page
		case "auth.admin.sessions.inspect":
			public = sessionSummary
		case "auth.admin.accounts.suspend", "auth.admin.accounts.reactivate":
			suspend := request.Operation == "auth.admin.accounts.suspend"
			mutation = func(db sqliteExecutor) error {
				updated, err := db.Exec(`UPDATE plugin_state SET value_json=json_set(value_json,'$.suspended',json(?)) WHERE project_id=? AND namespace=? AND state_key=? AND json_extract(value_json,'$.userId')=?`, map[bool]string{true: "true", false: "false"}[suspend], project, authAccountsNamespace, username, project+":"+username)
				if err != nil {
					return err
				}
				count, err := updated.RowsAffected()
				if err != nil {
					return err
				}
				if count != 1 {
					return ErrStoredResourceMissing
				}
				var revoked int64
				if suspend {
					deleted, err := db.Exec(`DELETE FROM plugin_state WHERE project_id=? AND namespace=? AND json_extract(value_json,'$.username')=?`, project, authSessionsNamespace, username)
					if err != nil {
						return err
					}
					revoked, err = deleted.RowsAffected()
					if err != nil {
						return err
					}
				}
				public = map[string]any{"username": username, "suspended": suspend, "revoked": revoked}
				return nil
			}
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
