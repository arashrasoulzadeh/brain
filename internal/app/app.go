package app

import (
	"net/http"

	"brain/internal/admin"
	"brain/internal/operator"
)

// Server owns the HTTP routes for the application process.
type Server struct {
	mux *http.ServeMux
}

// NewServer creates the smallest usable application server.
func NewServer() *Server {
	return NewServerWithAdmin(nil)
}

func NewServerWithAdmin(adminService *admin.LifecycleService) *Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /livez", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("ok\n"))
	})
	if adminService != nil {
		mux.Handle("POST /admin/actions", AdminHandler(adminService))
	}
	return &Server{mux: mux}
}

func NewServerWithOperatorAdmin(adminService *admin.LifecycleService, sessions *operator.Store, digest func(string) string) *Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /livez", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("ok\n"))
	})
	if adminService != nil && sessions != nil {
		mux.Handle("POST /admin/session", OperatorSessionHandler(sessions, digest))
		mux.Handle("DELETE /admin/session", OperatorLogoutHandler(sessions))
		mux.Handle("GET /admin/projects", AuthenticatedProjectsHandler(adminService, sessions))
		mux.Handle("GET /admin/plugins", AuthenticatedPluginsHandler(adminService, sessions))
		mux.Handle("GET /admin/audit", AuthenticatedAuditHandler(adminService, sessions))
		mux.Handle("GET /admin/system", AuthenticatedSystemHandler(adminService, sessions))
		mux.Handle("GET /admin/functions", AuthenticatedFunctionsHandler(adminService, sessions))
		mux.Handle("POST /admin/functions", AuthenticatedFunctionsHandler(adminService, sessions))
		mux.Handle("GET /admin/function-logs", AuthenticatedFunctionLogsHandler(adminService, sessions))
		mux.Handle("/public/{project_id}/{function_name}", PublicFunctionHandler(adminService))
		mux.Handle("POST /admin/actions", AuthenticatedAdminHandler(adminService, sessions))
	}
	return &Server{mux: mux}
}

// Handler returns the server's HTTP handler for tests and command wiring.
func (s *Server) Handler() http.Handler {
	return withAPIEnvelope(s.mux)
}
