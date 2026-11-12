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
	mux.HandleFunc("POST /admin/plugins/install", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "invalid form", http.StatusBadRequest)
			return
		}
		if err := admin.InstallPlugin(r.FormValue("project"), r.FormValue("plugin"), r.FormValue("version")); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
	})
	mux.HandleFunc("POST /admin/plugins/grant", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "invalid form", http.StatusBadRequest)
			return
		}
		if err := admin.GrantPlugin(r.FormValue("project"), r.FormValue("plugin"), r.FormValue("version"), r.FormValue("capability")); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("POST /admin/plugins/enable", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "invalid form", http.StatusBadRequest)
			return
		}
		if err := admin.EnablePlugin(r.FormValue("project"), r.FormValue("plugin"), r.FormValue("version")); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	return mux
}
