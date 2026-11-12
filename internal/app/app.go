package app

import "net/http"

// Server owns the HTTP routes for the application process.
type Server struct {
	mux *http.ServeMux
}

// NewServer creates the smallest usable application server.
func NewServer() *Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /livez", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("ok\n"))
	})
	return &Server{mux: mux}
}

// Handler returns the server's HTTP handler for tests and command wiring.
func (s *Server) Handler() http.Handler {
	return s.mux
}
