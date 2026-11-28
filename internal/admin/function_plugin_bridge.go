package admin

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"slices"
	"strings"

	"brain/internal/plugin"
)

const maxFunctionBridgeBytes = 256 * 1024
const maxFunctionPluginCalls = 16

type functionPluginRequest struct {
	Operation string          `json:"operation"`
	Payload   json.RawMessage `json:"payload"`
}

type functionPluginResponse struct {
	Result json.RawMessage `json:"result,omitempty"`
	Error  string          `json:"error,omitempty"`
}

// Only the trusted host chooses project and principal. The function can select
// an opted-in operation and its arguments, not credentials or routing identity.
func (s *LifecycleService) authorizeFunctionPluginRead(ctx context.Context, projectID, name, principal, operation, capability string) error {
	if ctx.Err() != nil || s.registry == nil || operation != "functions.invoke" || !strings.HasPrefix(principal, "operator:") {
		return plugin.ErrActorDenied
	}
	if s.projects != nil {
		if _, err := s.projects.Resolve(projectID); err != nil {
			return err
		}
	}
	if err := s.registry.AuthorizeFunctionBridgeSource(projectID, principal, strings.TrimSpace(name)); err != nil {
		return err
	}
	// Unlike the legacy draft cache, this privileged lookup fails closed when
	// the persisted grants are unavailable or corrupt.
	functionDrafts.Lock()
	defer functionDrafts.Unlock()
	data, err := os.ReadFile(functionDraftsPath)
	if err != nil {
		return plugin.ErrGrantDenied
	}
	var drafts []FunctionDraft
	if json.Unmarshal(data, &drafts) != nil {
		return plugin.ErrGrantDenied
	}
	for _, draft := range drafts {
		if draft.ProjectID == projectID && draft.Name == strings.TrimSpace(name) && slices.Contains(draft.Permissions, capability) {
			return nil
		}
	}
	return plugin.ErrGrantDenied
}

func (s *LifecycleService) functionPluginRead(ctx context.Context, projectID, name, principal, operation string, call functionPluginRequest) functionPluginResponse {
	if s.registry == nil || ctx.Err() != nil || operation != "functions.invoke" || !strings.HasPrefix(principal, "operator:") {
		return functionPluginResponse{Error: "Plugin calls require an operator sandbox test"}
	}
	if call.Operation == "" || len(call.Operation) > 128 {
		return functionPluginResponse{Error: "Invalid plugin operation"}
	}
	allow := func(capability string) error {
		return s.authorizeFunctionPluginRead(ctx, projectID, name, principal, operation, capability)
	}
	result, err := s.registry.InvokeFunctionReadContext(ctx, projectID, principal, call.Operation, call.Payload, allow)
	if err != nil {
		if errors.Is(err, plugin.ErrStoredResourceMissing) {
			return functionPluginResponse{Error: "Plugin resource not found"}
		}
		return functionPluginResponse{Error: "Plugin call denied or failed; check the operation, saved function permission, plugin installation and resource policy"}
	}
	if !json.Valid(result.Payload) {
		return functionPluginResponse{Error: "Plugin returned invalid JSON"}
	}
	return functionPluginResponse{Result: json.RawMessage(result.Payload)}
}

func readFunctionBridgeFrame(reader io.Reader) ([]byte, error) {
	var header [4]byte
	if _, err := io.ReadFull(reader, header[:]); err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint32(header[:])
	if length == 0 || length > maxFunctionBridgeBytes {
		return nil, errors.New("invalid plugin bridge frame")
	}
	payload := make([]byte, int(length))
	_, err := io.ReadFull(reader, payload)
	return payload, err
}

func writeFunctionBridgeFrame(writer io.Writer, payload []byte) error {
	if len(payload) == 0 || len(payload) > maxFunctionBridgeBytes {
		return errors.New("invalid plugin bridge response")
	}
	var header [4]byte
	binary.BigEndian.PutUint32(header[:], uint32(len(payload)))
	if n, err := writer.Write(header[:]); err != nil {
		return err
	} else if n != len(header) {
		return io.ErrShortWrite
	}
	if n, err := writer.Write(payload); err != nil {
		return err
	} else if n != len(payload) {
		return io.ErrShortWrite
	}
	return nil
}

// Each invocation has dedicated anonymous pipes and a bounded request budget.
// The outer Node runner owns fd 3 (request) and fd 4 (response).
func (s *LifecycleService) runFunctionCommand(ctx context.Context, command *exec.Cmd, projectID, name, principal, operation string) ([]byte, error) {
	requestReader, requestWriter, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	defer requestReader.Close()
	defer requestWriter.Close()
	responseReader, responseWriter, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	defer responseReader.Close()
	defer responseWriter.Close()
	bridgeCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	command.ExtraFiles = []*os.File{requestWriter, responseReader}
	go func() {
		defer responseWriter.Close()
		for count := 0; count <= maxFunctionPluginCalls; count++ {
			if bridgeCtx.Err() != nil {
				return
			}
			frame, err := readFunctionBridgeFrame(requestReader)
			if err != nil {
				return
			}
			response := functionPluginResponse{Error: "Invalid plugin call"}
			var call functionPluginRequest
			if count == maxFunctionPluginCalls {
				response.Error = "Plugin call limit exceeded"
			} else if json.Unmarshal(frame, &call) == nil {
				response = s.functionPluginRead(bridgeCtx, projectID, name, principal, operation, call)
			}
			if bridgeCtx.Err() != nil {
				return
			}
			encoded, err := json.Marshal(response)
			if err != nil || len(encoded) > maxFunctionBridgeBytes {
				encoded = []byte(`{"error":"Plugin response exceeds bridge limit"}`)
			}
			if writeFunctionBridgeFrame(responseWriter, encoded) != nil {
				return
			}
		}
	}()
	return command.Output()
}
