package admin

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
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
var functionNamePattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]{0,63}$`)

const maxFunctionSourceBytes = 256 * 1024

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

const functionDraftsPath = ".local/function-drafts.json"

func loadFunctionDraftsLocked() {
	data, err := os.ReadFile(functionDraftsPath)
	if err != nil {
		return
	}
	var drafts []FunctionDraft
	if json.Unmarshal(data, &drafts) != nil {
		return
	}
	for _, draft := range drafts {
		functionDrafts.items[draft.ProjectID+":"+draft.Name] = draft
	}
}

func persistFunctionDraftsLocked() error {
	if err := os.MkdirAll(filepath.Dir(functionDraftsPath), 0o750); err != nil {
		return err
	}
	drafts := make([]FunctionDraft, 0, len(functionDrafts.items))
	for _, draft := range functionDrafts.items {
		drafts = append(drafts, draft)
	}
	data, err := json.MarshalIndent(drafts, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(functionDraftsPath, data, 0o640)
}

func (s *LifecycleService) CreateFunctionDraft(actor operator.Actor, projectID, name, source string) (FunctionDraft, error) {
	if actor.Kind != operator.ActorOperator || strings.TrimSpace(projectID) == "" || !functionNamePattern.MatchString(strings.TrimSpace(name)) || strings.TrimSpace(source) == "" || len(source) > maxFunctionSourceBytes {
		return FunctionDraft{}, ErrFunctionDraftInvalid
	}
	if s.projects != nil {
		if _, err := s.projects.Resolve(projectID); err != nil {
			return FunctionDraft{}, err
		}
	}
	draft := FunctionDraft{ProjectID: projectID, Name: strings.TrimSpace(name), Source: source, CreatedAt: time.Now().UTC()}
	functionDrafts.Lock()
	loadFunctionDraftsLocked()
	functionDrafts.items[projectID+":"+draft.Name] = draft
	if err := persistFunctionDraftsLocked(); err != nil {
		functionDrafts.Unlock()
		return FunctionDraft{}, err
	}
	functionDrafts.Unlock()
	return draft, nil
}

func (s *LifecycleService) ListFunctionDrafts(actor operator.Actor, projectID string) ([]FunctionDraft, error) {
	if actor.Kind != operator.ActorOperator || strings.TrimSpace(projectID) == "" {
		return nil, ErrFunctionDraftInvalid
	}
	if s.projects != nil {
		if _, err := s.projects.Resolve(projectID); err != nil {
			return nil, err
		}
	}
	functionDrafts.Lock()
	defer functionDrafts.Unlock()
	loadFunctionDraftsLocked()
	drafts := make([]FunctionDraft, 0)
	for _, draft := range functionDrafts.items {
		if draft.ProjectID == projectID {
			drafts = append(drafts, draft)
		}
	}
	return drafts, nil
}

func (s *LifecycleService) TestFunctionDraft(actor operator.Actor, projectID, name, source string) (map[string]any, error) {
	if actor.Kind != operator.ActorOperator || strings.TrimSpace(projectID) == "" || !functionNamePattern.MatchString(strings.TrimSpace(name)) || strings.TrimSpace(source) == "" || len(source) > maxFunctionSourceBytes {
		return nil, ErrFunctionDraftInvalid
	}
	if s.projects != nil {
		if _, err := s.projects.Resolve(projectID); err != nil {
			return nil, err
		}
	}
	if !functionExportPattern.MatchString(source) || functionUnsafePattern.MatchString(source) {
		return nil, ErrFunctionSourceInvalid
	}
	return map[string]any{"valid": true, "project_id": projectID, "name": strings.TrimSpace(name), "source_bytes": len(source), "message": "source passed the host preflight and is ready for sandbox compilation"}, nil
}

func (s *LifecycleService) RunFunctionDraft(actor operator.Actor, projectID, name, source string, input json.RawMessage) (map[string]any, error) {
	if _, err := s.TestFunctionDraft(actor, projectID, name, source); err != nil {
		return nil, err
	}
	if len(input) == 0 {
		input = json.RawMessage(`null`)
	}
	if !json.Valid(input) {
		return nil, errors.New("function input must be valid JSON")
	}
	const runner = `const vm=await import("node:vm");const source=process.env.BRAIN_FUNCTION_SOURCE;const input=JSON.parse(process.env.BRAIN_FUNCTION_INPUT);const context=vm.createContext(Object.create(null));const module=new vm.SourceTextModule(source,{context});await module.link(()=>{throw new Error("imports are not permitted")});await module.evaluate({timeout:450});const handler=module.namespace.default;if(typeof handler!=="function")throw new Error("default export must be a function");const result=await Promise.race([Promise.resolve(handler(input)),new Promise((_,reject)=>setTimeout(()=>reject(new Error("function execution timed out")),400))]);process.stdout.write(JSON.stringify(result??null));`
	commandContext, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	command := exec.CommandContext(commandContext, "node", "--experimental-vm-modules", "--input-type=module", "-e", runner)
	command.Env = append(os.Environ(), "BRAIN_FUNCTION_SOURCE="+source, "BRAIN_FUNCTION_INPUT="+string(input))
	output, err := command.Output()
	if err != nil {
		return nil, errors.New("function execution failed")
	}
	var result any
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, errors.New("function returned invalid JSON")
	}
	return map[string]any{"valid": true, "project_id": projectID, "name": name, "result": result}, nil
}
