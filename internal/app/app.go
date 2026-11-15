package app

import "net/http"

import "brain/internal/admin"

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

// Handler returns the server's HTTP handler for tests and command wiring.
func (s *Server) Handler() http.Handler {
	return s.mux
}
