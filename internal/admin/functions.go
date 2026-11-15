package admin

import (
	"errors"
	"regexp"
	"strings"
	"sync"
	"time"

	"brain/internal/operator"
)

var ErrFunctionDraftInvalid = errors.New("function draft requires a name and source")
var ErrFunctionSourceInvalid = errors.New("function source must export a handler and may not use host or network access")

var functionExportPattern = regexp.MustCompile(`(?m)\bexport\s+default\b`)
var functionUnsafePattern = regexp.MustCompile(`(?m)\b(import|require|process|child_process|fs|net|fetch|Deno)\b`)

type FunctionDraft struct {
	ProjectID string    `json:"project_id"`
	Name      string    `json:"name"`
	Source    string    `json:"source"`
	CreatedAt time.Time `json:"created_at"`
}

var functionDrafts = struct {
	sync.RWMutex
	items map[string]FunctionDraft
}{items: make(map[string]FunctionDraft)}

func (s *LifecycleService) CreateFunctionDraft(actor operator.Actor, projectID, name, source string) (FunctionDraft, error) {
	if actor.Kind != operator.ActorOperator || strings.TrimSpace(projectID) == "" || strings.TrimSpace(name) == "" || strings.TrimSpace(source) == "" {
		return FunctionDraft{}, ErrFunctionDraftInvalid
	}
	draft := FunctionDraft{ProjectID: projectID, Name: strings.TrimSpace(name), Source: source, CreatedAt: time.Now().UTC()}
	functionDrafts.Lock()
	functionDrafts.items[projectID+":"+draft.Name] = draft
	functionDrafts.Unlock()
	return draft, nil
}

func (s *LifecycleService) ListFunctionDrafts(actor operator.Actor, projectID string) ([]FunctionDraft, error) {
	if actor.Kind != operator.ActorOperator || strings.TrimSpace(projectID) == "" {
		return nil, ErrFunctionDraftInvalid
	}
	functionDrafts.RLock()
	defer functionDrafts.RUnlock()
	drafts := make([]FunctionDraft, 0)
	for _, draft := range functionDrafts.items {
		if draft.ProjectID == projectID {
			drafts = append(drafts, draft)
		}
	}
	return drafts, nil
}

func (s *LifecycleService) TestFunctionDraft(actor operator.Actor, projectID, name, source string) (map[string]any, error) {
	if actor.Kind != operator.ActorOperator || strings.TrimSpace(projectID) == "" || strings.TrimSpace(name) == "" || strings.TrimSpace(source) == "" {
		return nil, ErrFunctionDraftInvalid
	}
	if !functionExportPattern.MatchString(source) || functionUnsafePattern.MatchString(source) {
		return nil, ErrFunctionSourceInvalid
	}
	return map[string]any{"valid": true, "project_id": projectID, "name": strings.TrimSpace(name), "source_bytes": len(source), "message": "source passed the host preflight and is ready for sandbox compilation"}, nil
}
