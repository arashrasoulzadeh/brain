package project

import "net/http"

type Resolver struct {
	Store *Store
}

func (r Resolver) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if r.Store == nil {
			http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
			return
		}
		id := req.Header.Get("X-Project-ID")
		caller, trusted := CallerFromContext(req.Context())
		if !trusted || caller.ProjectID != id {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		p, err := r.Store.Resolve(id)
		if err != nil {
			status := http.StatusNotFound
			if err == ErrInactive {
				status = http.StatusForbidden
			}
			http.Error(w, http.StatusText(status), status)
			return
		}
		next.ServeHTTP(w, req.WithContext(WithProject(req.Context(), p)))
	})
}
