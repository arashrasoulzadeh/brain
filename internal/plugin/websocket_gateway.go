package plugin

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"brain/internal/project"

	"github.com/coder/websocket"
)

const socketProtocol = "brain.v1"
const socketReadLimit = 64 << 10

type webSocketGateway struct {
	registry *Registry
	tickets  *WebSocketTicketStore
	slots    chan struct{}
}

// NewWebSocketGateway is constructed once, before the HTTP server begins
// serving requests. No plugin receives session tokens or raw socket handles.
func NewWebSocketGateway(registry *Registry) http.Handler {
	if registry == nil {
		return WebSocketEndpoint()
	}
	if registry.webSockets == nil {
		registry.BindWebSocketRegistry(NewWebSocketRegistry())
	}
	return &webSocketGateway{
		registry: registry,
		tickets: NewWebSocketTicketStore(4096, func(call HostCall, _ time.Time) error {
			return registry.authorizeActorAccess(call)
		}),
		slots: make(chan struct{}, 256),
	}
}

func (g *webSocketGateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	namespace, ok := websocketPath(r.URL.Path)
	p, resolved := project.FromContext(r.Context())
	if !ok || !resolved || !p.Active || (p.ID != namespace && p.Slug != namespace) || r.URL.RawQuery != "" {
		http.Error(w, "project boundary denied", http.StatusForbidden)
		return
	}
	if identity := r.Header.Get("X-Project-ID"); identity != "" && identity != p.ID && identity != p.Slug {
		http.Error(w, "project boundary denied", http.StatusForbidden)
		return
	}
	if r.Method == http.MethodGet && !isWebSocketUpgrade(r) {
		http.Error(w, "websocket upgrade required", http.StatusUpgradeRequired)
		return
	}
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}
	origin := r.Header.Get("Origin")
	if !sameSocketOrigin(r, origin) {
		http.Error(w, "origin denied", http.StatusForbidden)
		return
	}
	if r.Method == http.MethodPost {
		g.issue(w, r, p.ID, origin)
		return
	}
	g.connect(w, r, p.ID, origin)
}

// Same-host HTTPS (or loopback HTTP for native development) only. Deployments
// behind a TLS proxy may forward the public Host; no client forwarded headers
// are trusted to choose an origin.
func sameSocketOrigin(r *http.Request, origin string) bool {
	if len(r.Header.Values("Origin")) != 1 {
		return false
	}
	u, err := url.Parse(origin)
	if err != nil || u.Host != r.Host || u.User != nil || u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
		return false
	}
	return u.Scheme == "https" || (u.Scheme == "http" && (u.Hostname() == "localhost" || u.Hostname() == "127.0.0.1" || u.Hostname() == "::1"))
}

func socketBearer(r *http.Request) string {
	if len(r.Header.Values("Authorization")) != 1 {
		return ""
	}
	scheme, token, ok := strings.Cut(r.Header.Get("Authorization"), " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") || len(token) > 1024 || strings.ContainsAny(token, " \t\r\n") {
		return ""
	}
	return token
}

func (g *webSocketGateway) issue(w http.ResponseWriter, r *http.Request, projectID, origin string) {
	token := socketBearer(r)
	if token == "" {
		http.Error(w, "user session required", http.StatusUnauthorized)
		return
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
	decoder.DisallowUnknownFields()
	var input struct {
		Path string `json:"path"`
	}
	if decoder.Decode(&input) != nil || decoder.Decode(new(any)) != io.EOF ||
		!strings.HasPrefix(input.Path, "/") || strings.ContainsAny(input.Path, "%\\?#") {
		http.Error(w, "invalid socket route", http.StatusBadRequest)
		return
	}
	actor, err := g.registry.ResolveUserSession(projectID, token)
	match, found := g.registry.MatchHTTPRoute(projectID, http.MethodPost, input.Path)
	if err != nil || !found || match.Auth != "user" || match.WebSocketAdmission == "" {
		http.Error(w, "connection denied", http.StatusForbidden)
		return
	}
	binding := webSocketTicketBinding{call: HostCall{
		ProjectID: projectID, ActorID: actor.ID, PluginID: match.PluginID, Version: match.Version,
		Operation: match.Operation, Capability: match.Capability, Resource: match.Params[match.ResourceField],
	}, session: token, path: input.Path, origin: origin}
	if _, err := g.admit(binding); err != nil {
		http.Error(w, "connection denied", http.StatusForbidden)
		return
	}
	ticket, err := g.tickets.issue(binding, time.Now(), MaxWebSocketTicketLifetime)
	if err != nil {
		http.Error(w, "connection denied", http.StatusForbidden)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"ticket": ticket, "expires_in": int(MaxWebSocketTicketLifetime.Seconds()), "protocol": socketProtocol})
}

func (g *webSocketGateway) admit(binding webSocketTicketBinding) (HTTPRouteMatch, error) {
	r := g.registry
	call := binding.call
	actor, err := r.ResolveUserSession(call.ProjectID, binding.session)
	if err != nil || actor.ID != call.ActorID {
		return HTTPRouteMatch{}, ErrActorDenied
	}
	match, ok := r.MatchHTTPRoute(call.ProjectID, http.MethodPost, binding.path)
	if !ok || match.PluginID != call.PluginID || match.Version != call.Version || match.Operation != call.Operation ||
		match.Capability != call.Capability || match.Params[match.ResourceField] != call.Resource ||
		match.Auth != "user" || match.WebSocketAdmission == "" {
		return HTTPRouteMatch{}, ErrActorDenied
	}
	r.readinessMu.RLock()
	manifest := r.registered[match.PluginID+"@"+match.Version].Manifest
	admission, found := manifest.Route(match.WebSocketAdmission)
	r.readinessMu.RUnlock()
	if !found || !isValidProjectID(call.Resource) {
		return HTTPRouteMatch{}, ErrActorDenied
	}
	admissionPath := mountedHTTPPath(manifest.Mount, strings.ReplaceAll(admission.Path, "{"+match.ResourceField+"}", call.Resource))
	admissionMatch, found := r.MatchHTTPRoute(call.ProjectID, http.MethodPost, admissionPath)
	if !found || admissionMatch.PluginID != call.PluginID || admissionMatch.Version != call.Version || admissionMatch.Operation != admission.Operation {
		return HTTPRouteMatch{}, ErrActorDenied
	}
	payload, _ := json.Marshal(map[string]string{match.ResourceField: call.Resource})
	request := WorkerRequest{Protocol: r.HostAPI, ProjectID: call.ProjectID, PluginID: call.PluginID,
		Version: call.Version, Operation: call.Operation, Payload: payload, httpRoute: &match,
		accessPrincipal: "user:" + actor.ID, requireAccessPolicy: true}
	if err := r.authorizeInvocation(request, false); err != nil {
		return HTTPRouteMatch{}, err
	}
	response, err := r.InvokeHTTPRoute(call.ProjectID, admissionMatch, payload, binding.session)
	var admitted struct {
		Admitted bool `json:"admitted"`
	}
	if err != nil || !response.Success || response.ErrorCode != "" || json.Unmarshal(response.Payload, &admitted) != nil || !admitted.Admitted {
		return HTTPRouteMatch{}, ErrActorDenied
	}
	return match, nil
}

func socketTicket(r *http.Request) string {
	protocols := strings.Split(strings.Join(r.Header.Values("Sec-WebSocket-Protocol"), ","), ",")
	if len(protocols) != 2 || strings.TrimSpace(protocols[0]) != socketProtocol {
		return ""
	}
	ticket, ok := strings.CutPrefix(strings.TrimSpace(protocols[1]), "ticket.")
	if !ok {
		return ""
	}
	return ticket
}

type trackedSocket struct{ connection *websocket.Conn }

func (s trackedSocket) Close() error { return s.connection.CloseNow() }

func (g *webSocketGateway) connect(w http.ResponseWriter, r *http.Request, projectID, origin string) {
	token := socketTicket(r)
	if token == "" {
		http.Error(w, "connection ticket required", http.StatusUnauthorized)
		return
	}
	select {
	case g.slots <- struct{}{}:
		defer func() { <-g.slots }()
	default:
		http.Error(w, "socket capacity exhausted", http.StatusServiceUnavailable)
		return
	}
	binding, err := g.tickets.consume(projectID, token, time.Now())
	if err != nil || binding.origin != origin {
		http.Error(w, "connection denied", http.StatusUnauthorized)
		return
	}
	if _, err := g.admit(binding); err != nil {
		http.Error(w, "connection denied", http.StatusForbidden)
		return
	}
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{Subprotocols: []string{socketProtocol}})
	if err != nil {
		return
	}
	defer conn.CloseNow()
	// Use a non-secret identifier for lifecycle tracking. Each consumed ticket
	// creates at most one connection.
	id, err := newModelRevision()
	if err != nil || g.registry.webSockets.RegisterProject(id, projectID, trackedSocket{conn}) != nil {
		return
	}
	defer g.registry.webSockets.Unregister(id)
	conn.SetReadLimit(socketReadLimit)
	// Close on expiry or inactivity; every message also rechecks session,
	// project, version, resource and worker membership before invocation.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	for {
		readCtx, done := context.WithTimeout(ctx, 30*time.Second)
		kind, data, err := conn.Read(readCtx)
		done()
		if err != nil {
			return
		}
		match, err := g.admit(binding)
		if err != nil || kind != websocket.MessageText {
			_ = conn.Close(websocket.StatusPolicyViolation, "connection denied")
			return
		}
		var fields map[string]json.RawMessage
		if json.Unmarshal(data, &fields) != nil || fields == nil {
			_ = conn.Close(websocket.StatusInvalidFramePayloadData, "invalid message")
			return
		}
		for key := range fields {
			if key == match.ResourceField || key == "projectId" || key == "project_id" || key == "project" || strings.HasPrefix(key, "_brain") {
				_ = conn.Close(websocket.StatusPolicyViolation, "reserved field")
				return
			}
		}
		fields[match.ResourceField], _ = json.Marshal(binding.call.Resource)
		payload, err := json.Marshal(fields)
		if err != nil {
			return
		}
		response, err := g.registry.InvokeHTTPRoute(projectID, match, payload, binding.session)
		if err != nil || !response.Success || response.ErrorCode != "" || !json.Valid(response.Payload) || len(response.Payload) > socketReadLimit {
			_ = conn.Close(websocket.StatusPolicyViolation, "message rejected")
			return
		}
		writeCtx, done := context.WithTimeout(ctx, 5*time.Second)
		err = conn.Write(writeCtx, websocket.MessageText, response.Payload)
		done()
		if err != nil {
			return
		}
	}
}
