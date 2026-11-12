package dashboard

import (
	"net/http"
	"strings"
)

func Handler(admin ProjectAdmin) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /admin/projects", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "invalid form", http.StatusBadRequest)
			return
		}
		id := strings.TrimSpace(r.FormValue("id"))
		if err := admin.CreateProject(id); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
	})
	return mux
}
