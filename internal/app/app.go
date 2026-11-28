package app

import (
	"context"
	"net/http"
	"time"

	"brain/internal/admin"
	"brain/internal/cortex"
	"brain/internal/operator"
	"brain/internal/project"
)

// Server owns the HTTP routes for the application process.
type Server struct {
	mux             *http.ServeMux
	readiness       func(context.Context) error
	projects        *project.Store
	neuronScheduler *cortex.Scheduler
}

func newServer() *Server {
	s := &Server{mux: http.NewServeMux()}
	s.mux.HandleFunc("GET /_system/livez", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("ok\n"))
	})
	s.mux.HandleFunc("GET /_system/readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if s.readiness == nil || s.readiness(ctx) != nil {
			http.Error(w, "not ready", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("ok\n"))
	})
	return s
}

// SetReadinessCheck configures dependency checks before the server starts.
func (s *Server) SetReadinessCheck(check func(context.Context) error) {
	s.readiness = check
}

// NewServer creates the smallest usable application server.
func NewServer() *Server {
	return NewServerWithAdmin(nil)
}

func NewServerWithAdmin(adminService *admin.LifecycleService) *Server {
	server := newServer()
	if adminService != nil {
		server.projects = adminService.ProjectStore()
		server.mux.Handle("POST /_system/admin/actions", adminNamespaceHandler(AdminHandler(adminService), true))
		server.mux.Handle("POST /{project}/admin/actions", ProjectCommandHandler(adminNamespaceHandler(AdminHandler(adminService), false)))
	}
	return server
}

func NewServerWithOperatorAdmin(adminService *admin.LifecycleService, sessions *operator.Store, digest func(string) string) *Server {
	server := newServer()
	if adminService == nil || sessions == nil {
		return server
	}
	server.projects = adminService.ProjectStore()
	mux := server.mux
	mux.Handle("POST /_system/admin/session", OperatorSessionHandler(sessions, digest))
	mux.Handle("DELETE /_system/admin/session", OperatorLogoutHandler(sessions))
	mux.Handle("GET /_system/admin/projects", AuthenticatedProjectsHandler(adminService, sessions))
	mux.Handle("GET /_system/admin/plugins", AuthenticatedPluginsHandler(adminService, sessions))
	mux.Handle("GET /_system/admin/audit", AuthenticatedAuditHandler(adminService, sessions))
	mux.Handle("GET /_system/admin/system", AuthenticatedSystemHandler(adminService, sessions))
	mux.Handle("GET /_system/admin/neurons", server.neuronStatusHandler(sessions))
	mux.Handle("POST /_system/admin/actions", adminNamespaceHandler(AuthenticatedAdminHandler(adminService, sessions), true))

	mux.Handle("POST /{project}/admin/actions", AuthenticatedProjectCommandHandler(adminNamespaceHandler(AuthenticatedAdminHandler(adminService, sessions), false), sessions))
	mux.Handle("GET /{project}/admin/plugins", ProjectQueryHandler(AuthenticatedPluginsHandler(adminService, sessions)))
	mux.Handle("GET /{project}/admin/functions", ProjectQueryHandler(AuthenticatedFunctionsHandler(adminService, sessions)))
	mux.Handle("POST /{project}/admin/functions", AuthenticatedProjectCommandHandler(AuthenticatedFunctionsHandler(adminService, sessions), sessions))
	mux.Handle("GET /{project}/admin/function-logs", ProjectQueryHandler(AuthenticatedFunctionLogsHandler(adminService, sessions)))
	mux.Handle("GET /{project}/admin/routes", AuthenticatedProjectRoutesHandler(adminService, sessions))
	mux.Handle("POST /{project}/admin/files/blob", AuthenticatedFileTransferHandler(adminService, sessions))
	mux.Handle("GET /{project}/admin/files/blob", AuthenticatedFileTransferHandler(adminService, sessions))
	mux.Handle("/{project}/functions/{function_name}", PublicFunctionHandler(adminService))
	socket := adminService.WebSocketHandler()
	mux.Handle("GET /{project}/ws", socket)
	mux.Handle("POST /{project}/ws", socket)
	mux.Handle("/{project}/{resource...}", PluginHTTPHandler(adminService, sessions))
	return server
}

// Handler returns the server's HTTP handler for tests and command wiring.
func (s *Server) Handler() http.Handler {
	return withAPIEnvelope((project.Resolver{Store: s.projects}).URLMiddleware(s.mux))
}
