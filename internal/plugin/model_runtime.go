package plugin

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math"
	"regexp"
)

var ErrModelConflict = errors.New("model record missing, already exists, or revision changed")
var ErrModelSchema = errors.New("model schema version mismatch")
var recordIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,128}$`)

type modelRecord struct {
	ID       string         `json:"id"`
	Revision string         `json:"revision"`
	State    string         `json:"state,omitempty"`
	Fields   map[string]any `json:"fields"`
}

type modelInput struct {
	ID       string         `json:"id"`
	Revision string         `json:"revision"`
	Fields   map[string]any `json:"fields"`
	After    string         `json:"after"`
}

func modelFieldsValid(spec ModelSpec, fields map[string]any) bool {
	if len(fields) != len(spec.Fields) {
		return false
	}
	for key, kind := range spec.Fields {
		switch v := fields[key].(type) {
		case string:
			if kind != "string" || len(v) > 1024 {
				return false
			}
		case bool:
			if kind != "boolean" {
				return false
			}
		case float64:
			if kind != "integer" || math.Trunc(v) != v || math.Abs(v) > 9007199254740991 {
				return false
			}
		default:
			return false
		}
	}
	return true
}

func newModelRevision() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}

// Model operations are host-only. Lifecycle and grants remain locked through
// the database operation; revisions prevent stale writes across host processes.
func (r *Registry) invokeModel(req WorkerRequest) (WorkerResponse, bool, error) {
	r.readinessMu.RLock()
	defer r.readinessMu.RUnlock()
	// Policy changes must not race a permitted invocation's final storage commit.
	// All registry authorization and commits hold the installation read lock.
	if req.PluginID == accessPluginID {
		r.installationMu.Lock()
		defer r.installationMu.Unlock()
	} else {
		r.installationMu.RLock()
		defer r.installationMu.RUnlock()
	}
	manifest := r.registered[req.PluginID+"@"+req.Version].Manifest
	route, ok := manifest.Route(req.Operation)
	if !ok || route.Helper != "host.model" {
		return WorkerResponse{}, false, nil
	}
	if err := r.authorizeInvocationLocked(req, false); err != nil {
		return WorkerResponse{}, true, err
	}
	installation := r.installations[req.ProjectID+":"+req.PluginID+"@"+req.Version]
	if resources := installation.Resources[route.Capability]; len(resources) > 0 && !resources[route.Model.Name] {
		return WorkerResponse{}, true, ErrGrantDenied
	}
	storage, ok := r.storage.(*SQLiteStorage)
	if !ok || storage == nil {
		return WorkerResponse{}, true, ErrLifecycleStorage
	}
	if req.PluginID == accessPluginID && route.Model.Name == accessRuleModel && route.Model.Action == "create" {
		if err := r.validateAccessRuleLocked(req); err != nil {
			return WorkerResponse{}, true, err
		}
	}
	result, err := storage.executeModel(req, manifest.Models[route.Model.Name], *route.Model)
	if err != nil {
		return WorkerResponse{}, true, err
	}
	if req.PluginID == accessPluginID && route.Model.Name == accessRuleModel && route.Model.Action != "read" && route.Model.Action != "list" {
		var input modelInput
		_ = json.Unmarshal(req.Payload, &input)
		r.appendAudit(AuditEvent{Action: "access_rule_" + route.Model.Action, ProjectID: req.ProjectID, PluginID: req.PluginID, Version: req.Version, Detail: input.ID})
	}
	data, err := json.Marshal(result)
	if err != nil {
		return WorkerResponse{}, true, err
	}
	return WorkerResponse{Protocol: req.Protocol, Success: true, Payload: data}, true, nil
}

func scanModel(row interface{ Scan(...any) error }, spec ModelSpec) (modelRecord, error) {
	var out modelRecord
	var version int
	var fields string
	if err := row.Scan(&out.ID, &version, &out.Revision, &out.State, &fields); err != nil {
		return out, err
	}
	if version != spec.Version {
		return out, ErrModelSchema
	}
	if err := json.Unmarshal([]byte(fields), &out.Fields); err != nil {
		return out, err
	}
	if !modelFieldsValid(spec, out.Fields) {
		return out, ErrModelSchema
	}
	return out, nil
}

func (s *SQLiteStorage) executeModel(req WorkerRequest, spec ModelSpec, route ModelRoute) (any, error) {
	var in modelInput
	if err := json.Unmarshal(req.Payload, &in); err != nil {
		return nil, ErrInvalidWorkerRequest
	}
	if route.Action == "list" {
		if in.After != "" && !recordIDPattern.MatchString(in.After) {
			return nil, ErrInvalidWorkerRequest
		}
		rows, err := s.db.Query(`SELECT record_id,schema_version,revision,state,fields_json FROM plugin_model_records WHERE project_id=? AND plugin_id=? AND model_name=? AND record_id>? ORDER BY record_id LIMIT 20`, req.ProjectID, req.PluginID, route.Name, in.After)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		records := []modelRecord{}
		for rows.Next() {
			record, err := scanModel(rows, spec)
			if err != nil {
				return nil, err
			}
			records = append(records, record)
		}
		return records, rows.Err()
	}
	if !recordIDPattern.MatchString(in.ID) {
		return nil, ErrInvalidWorkerRequest
	}
	read := func() (modelRecord, error) {
		record, err := scanModel(s.db.QueryRow(`SELECT record_id,schema_version,revision,state,fields_json FROM plugin_model_records WHERE project_id=? AND plugin_id=? AND model_name=? AND record_id=?`, req.ProjectID, req.PluginID, route.Name, in.ID), spec)
		if errors.Is(err, sql.ErrNoRows) {
			return record, ErrModelConflict
		}
		return record, err
	}
	if route.Action == "read" {
		return read()
	}
	if route.Action == "create" {
		if !modelFieldsValid(spec, in.Fields) {
			return nil, ErrInvalidWorkerRequest
		}
		revision, err := newModelRevision()
		if err != nil {
			return nil, err
		}
		encoded, err := json.Marshal(in.Fields)
		if err != nil {
			return nil, err
		}
		result, err := s.db.Exec(`INSERT INTO plugin_model_records(project_id,plugin_id,model_name,record_id,schema_version,revision,state,fields_json) VALUES(?,?,?,?,?,?,?,?) ON CONFLICT(project_id,plugin_id,model_name,record_id) DO NOTHING`, req.ProjectID, req.PluginID, route.Name, in.ID, spec.Version, revision, spec.Initial, string(encoded))
		if err != nil {
			return nil, err
		}
		n, err := result.RowsAffected()
		if err != nil {
			return nil, err
		}
		if n != 1 {
			return nil, ErrModelConflict
		}
		return modelRecord{ID: in.ID, Revision: revision, State: spec.Initial, Fields: in.Fields}, nil
	}
	current, err := read()
	if err != nil {
		return nil, err
	}
	if in.Revision == "" || current.Revision != in.Revision {
		return nil, ErrModelConflict
	}
	var result sql.Result
	if route.Action == "delete" {
		result, err = s.db.Exec(`DELETE FROM plugin_model_records WHERE project_id=? AND plugin_id=? AND model_name=? AND record_id=? AND schema_version=? AND revision=?`, req.ProjectID, req.PluginID, route.Name, in.ID, spec.Version, in.Revision)
	} else {
		if route.Action == "update" {
			if !modelFieldsValid(spec, in.Fields) {
				return nil, ErrInvalidWorkerRequest
			}
			current.Fields = in.Fields
		} else if route.Action == "transition" {
			transition := spec.Transitions[route.Transition]
			if current.State != transition.From {
				return nil, ErrModelConflict
			}
			current.State = transition.To
		} else {
			return nil, ErrInvalidWorkerRequest
		}
		current.Revision, err = newModelRevision()
		if err != nil {
			return nil, err
		}
		var encoded []byte
		encoded, err = json.Marshal(current.Fields)
		if err != nil {
			return nil, err
		}
		result, err = s.db.Exec(`UPDATE plugin_model_records SET revision=?,state=?,fields_json=? WHERE project_id=? AND plugin_id=? AND model_name=? AND record_id=? AND schema_version=? AND revision=?`, current.Revision, current.State, string(encoded), req.ProjectID, req.PluginID, route.Name, in.ID, spec.Version, in.Revision)
	}
	if err != nil {
		return nil, err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if n != 1 {
		return nil, ErrModelConflict
	}
	if route.Action == "delete" {
		return map[string]bool{"deleted": true}, nil
	}
	return current, nil
}
