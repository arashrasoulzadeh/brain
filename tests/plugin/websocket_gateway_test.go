package plugin_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"brain/internal/admin"
	"brain/internal/app"
	"brain/internal/operator"
	"brain/internal/plugin"
	"brain/internal/project"

	"github.com/coder/websocket"
)

func TestWebSocketGatewayBindsSessionMembershipResourceAndDurableSequence(t *testing.T) {
	auth, err := plugin.LoadManifest("../../plugins/auth/plugin.yaml")
	if err != nil {
		t.Fatal(err)
	}
	arena, err := plugin.LoadManifest("../../plugins/arena/plugin.yaml")
	if err != nil {
		t.Fatal(err)
	}
	r, storage := accessRegistry(t, auth, arena)
	projects := project.NewStore(project.Project{ID: "alpha", Active: true}, project.Project{ID: "beta", Active: true})
	r.BindProjectResolver(func(id string) (plugin.ProjectSnapshot, error) {
		model, err := projects.Resolve(id)
		return plugin.ProjectSnapshot{ID: model.ID, Active: model.Active}, err
	})
	seedVerifiedSession(t, storage, "alpha")
	authWorker := &storageCommitWorker{handle: func(req plugin.WorkerRequest) plugin.WorkerResponse {
		var input struct{ Token string }
		_ = json.Unmarshal(req.Payload, &input)
		if input.Token != "valid-session" || req.ProjectID != "alpha" {
			return plugin.WorkerResponse{Protocol: req.Protocol, ErrorCode: "invalid session"}
		}
		return plugin.WorkerResponse{Protocol: req.Protocol, Success: true, Payload: []byte(`{"userId":"alpha:alice"}`)}
	}}
	arenaWorker := &storageCommitWorker{handle: func(req plugin.WorkerRequest) plugin.WorkerResponse {
		var input struct {
			RoomID   string `json:"roomId"`
			PlayerID string `json:"playerId"`
			Sequence int    `json:"sequence"`
			State    struct {
				Players   []string
				Sequences map[string]int
			} `json:"_brainArena"`
		}
		if json.Unmarshal(req.Payload, &input) != nil || req.HostContext == nil || req.HostContext.Auth.Caller.ID != "alpha:alice" {
			return plugin.WorkerResponse{Protocol: req.Protocol, ErrorCode: "invalid actor"}
		}
		member := false
		for _, player := range input.State.Players {
			member = member || player == req.HostContext.Auth.Caller.ID
		}
		if !member {
			return plugin.WorkerResponse{Protocol: req.Protocol, ErrorCode: "not a member"}
		}
		if req.Operation == "arena.connect" {
			return plugin.WorkerResponse{Protocol: req.Protocol, Success: true, Payload: []byte(`{"admitted":true}`)}
		}
		if req.Operation != "arena.player.input" || input.PlayerID != "alpha:alice" || input.RoomID != "one" || input.Sequence <= input.State.Sequences[input.PlayerID] {
			return plugin.WorkerResponse{Protocol: req.Protocol, ErrorCode: "invalid input"}
		}
		return plugin.WorkerResponse{Protocol: req.Protocol, Success: true, Payload: []byte(`{"accepted":true}`)}
	}}
	installAccessTestPlugin(t, r, "alpha", auth, authWorker)
	installAccessTestPlugin(t, r, "alpha", arena, arenaWorker)
	installAccessTestPlugin(t, r, "alpha", accessManifest(t), &contractWorker{ready: true})
	identity, _ := auth.Route("auth.identity")
	connect, _ := arena.Route("arena.connect")
	input, _ := arena.Route("arena.player.input")
	createAccessRule(t, r, "alpha", "auth-host", "host", auth, identity, "", true)
	for _, room := range []string{"one", "two"} {
		players := []string{"alpha:bob", "alpha:carol"}
		if room == "one" {
			players[0] = "alpha:alice"
		}
		if err := storage.PutState("alpha", "arena", room, map[string]any{"state": "active", "players": players, "sequences": map[string]any{}}); err != nil {
			t.Fatal(err)
		}
	}
	// Verify rejected requests before any session or worker is consulted.
	gateway := plugin.NewWebSocketGateway(r)
	for _, tc := range []struct {
		name, method, path, origin, bearer, body, identity, protocol string
		want                                                         int
	}{
		{"wrong path", "GET", "/alpha/other", "", "", "", "", "", 403},
		{"query token", "GET", "/alpha/ws?ticket=secret", "", "", "", "", "", 403},
		{"wrong project", "POST", "/beta/ws", "", "", "", "", "", 403},
		{"wrong identity", "POST", "/alpha/ws", "", "", "", "beta", "", 403},
		{"not upgrade", "GET", "/alpha/ws", "", "", "", "", "", 426},
		{"wrong method", "DELETE", "/alpha/ws", "", "", "", "", "", 404},
		{"no origin", "POST", "/alpha/ws", "", "", "", "", "", 403},
		{"foreign origin", "POST", "/alpha/ws", "https://foreign.example", "", "", "", "", 403},
		{"insecure remote", "POST", "/alpha/ws", "http://example.test", "", "", "", "", 403},
		{"missing token", "POST", "/alpha/ws", "https://example.test", "", "", "", "", 401},
		{"wrong auth scheme", "POST", "/alpha/ws", "https://example.test", "Basic abc", "", "", "", 401},
		{"invalid JSON", "POST", "/alpha/ws", "https://example.test", "Bearer abc", "{", "", "", 400},
		{"unknown field", "POST", "/alpha/ws", "https://example.test", "Bearer abc", `{"path":"/arena/one/input","extra":1}`, "", "", 400},
		{"relative route", "POST", "/alpha/ws", "https://example.test", "Bearer abc", `{"path":"arena/one/input"}`, "", "", 400},
		{"encoded route", "POST", "/alpha/ws", "https://example.test", "Bearer abc", `{"path":"/arena/%2e/input"}`, "", "", 400},
		{"missing ticket", "GET", "/alpha/ws", "https://example.test", "", "", "", "brain.v1", 401},
		{"invalid ticket protocol", "GET", "/alpha/ws", "https://example.test", "", "", "", "brain.v1, bad.abc", 401},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, "https://example.test"+tc.path, strings.NewReader(tc.body))
			req = req.WithContext(project.WithProject(req.Context(), project.Project{ID: "alpha", Active: true}))
			if tc.origin != "" {
				req.Header.Set("Origin", tc.origin)
			}
			if tc.bearer != "" {
				req.Header.Set("Authorization", tc.bearer)
			}
			if tc.identity != "" {
				req.Header.Set("X-Project-ID", tc.identity)
			}
			if tc.protocol != "" {
				req.Header.Set("Upgrade", "websocket")
				req.Header.Set("Connection", "Upgrade")
				req.Header.Set("Sec-WebSocket-Protocol", tc.protocol)
			}
			rec := httptest.NewRecorder()
			gateway.ServeHTTP(rec, req)
			if rec.Code != tc.want {
				t.Fatalf("status = %d, want %d", rec.Code, tc.want)
			}
		})
	}
	operators := &operator.Store{}
	server := httptest.NewServer(app.NewServerWithOperatorAdmin(admin.NewLifecycleServiceWithProjects(operators, r, projects), operators, func(value string) string { return value }).Handler())
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	issue := func(path, token, origin string, status int) string {
		t.Helper()
		payload, err := json.Marshal(map[string]string{"path": path})
		if err != nil {
			t.Fatal(err)
		}
		request, err := http.NewRequestWithContext(ctx, http.MethodPost, server.URL+"/alpha/ws", strings.NewReader(string(payload)))
		if err != nil {
			t.Fatal(err)
		}
		request.Header.Set("Authorization", "Bearer "+token)
		request.Header.Set("Origin", origin)
		request.Header.Set("Content-Type", "application/json")
		response, err := server.Client().Do(request)
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		if response.StatusCode != status {
			t.Fatalf("issue %s: %d, want %d", path, response.StatusCode, status)
		}
		if status != http.StatusOK {
			return ""
		}
		var result struct{ Ticket string }
		if err := json.NewDecoder(response.Body).Decode(&result); err != nil || result.Ticket == "" || response.Header.Get("Cache-Control") != "no-store" {
			t.Fatalf("invalid ticket response: %v", err)
		}
		return result.Ticket
	}
	dial := func(projectID, token string, status int) *websocket.Conn {
		t.Helper()
		conn, response, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http")+"/"+projectID+"/ws", &websocket.DialOptions{
			HTTPHeader: http.Header{"Origin": []string{server.URL}}, Subprotocols: []string{"brain.v1", "ticket." + token},
		})
		if status != http.StatusSwitchingProtocols {
			if response != nil {
				defer response.Body.Close()
			}
			if err == nil || response == nil || response.StatusCode != status {
				if conn != nil {
					_ = conn.CloseNow()
				}
				t.Fatalf("unexpected upgrade: %v, %v", response, err)
			}
			return nil
		}
		if err != nil || conn.Subprotocol() != "brain.v1" {
			t.Fatalf("upgrade failed: %v", err)
		}
		return conn
	}
	issue("/arena/one/input", "valid-session", server.URL, http.StatusForbidden)
	issue("/arena/one/input", "bad-token", server.URL, http.StatusForbidden)
	issue("/arena/one/input", "valid-session", "https://foreign.example", http.StatusForbidden)
	for _, room := range []string{"one", "two"} {
		createAccessRule(t, r, "alpha", "connect-"+room, "user:alpha:alice", arena, connect, room, true)
	}
	createAccessRule(t, r, "alpha", "input-two", "user:alpha:alice", arena, input, "two", true)
	revision := createAccessRule(t, r, "alpha", "input-one", "user:alpha:alice", arena, input, "one", true)
	issue("/arena/two/input", "valid-session", server.URL, http.StatusForbidden)
	issue("/arena/input", "valid-session", server.URL, http.StatusForbidden)
	ticket := issue("/arena/one/input", "valid-session", server.URL, http.StatusOK)
	dial("beta", ticket, http.StatusUnauthorized)
	conn := dial("alpha", ticket, http.StatusSwitchingProtocols)
	defer conn.CloseNow()
	dial("alpha", ticket, http.StatusUnauthorized)
	if err := conn.Write(ctx, websocket.MessageText, []byte(`{"sequence":1,"payload":{"move":"left"}}`)); err != nil {
		t.Fatal(err)
	}
	kind, data, err := conn.Read(ctx)
	if err != nil || kind != websocket.MessageText || string(data) != `{"accepted":true}` {
		t.Fatalf("worker response: %s (%v)", data, err)
	}
	state, err := storage.State("alpha", "arena", "one")
	if err != nil {
		t.Fatal(err)
	}
	sequences, ok := state["sequences"].(map[string]any)
	if !ok || sequences["alpha:alice"] != float64(1) {
		t.Fatalf("durable input not bound to actor: %v", state)
	}
	for _, tc := range []struct {
		name string
		kind websocket.MessageType
		body string
		want websocket.StatusCode
	}{
		{"binary", websocket.MessageBinary, "invalid", websocket.StatusPolicyViolation},
		{"null", websocket.MessageText, "null", websocket.StatusInvalidFramePayloadData},
		{"resource override", websocket.MessageText, `{"roomId":"two","sequence":2}`, websocket.StatusPolicyViolation},
		{"project override", websocket.MessageText, `{"projectId":"beta","sequence":2}`, websocket.StatusPolicyViolation},
		{"host field override", websocket.MessageText, `{"_brainArena":{},"sequence":2}`, websocket.StatusPolicyViolation},
	} {
		t.Run(tc.name, func(t *testing.T) {
			bad := dial("alpha", issue("/arena/one/input", "valid-session", server.URL, http.StatusOK), http.StatusSwitchingProtocols)
			defer bad.CloseNow()
			if err := bad.Write(ctx, tc.kind, []byte(tc.body)); err != nil {
				t.Fatal(err)
			}
			_, _, err := bad.Read(ctx)
			if websocket.CloseStatus(err) != tc.want {
				t.Fatalf("close status = %v, want %v", err, tc.want)
			}
		})
	}
	revokeAccessRule(t, r, "alpha", "input-one", revision)
	if err := conn.Write(ctx, websocket.MessageText, []byte(`{"sequence":2}`)); err == nil {
		if _, _, err := conn.Read(ctx); err == nil {
			t.Fatal("revoked socket accepted another input")
		}
	}
	issue("/arena/one/input", "valid-session", server.URL, http.StatusForbidden)
}
