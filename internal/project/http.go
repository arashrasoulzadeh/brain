package project

import "net/http"

type Resolver struct {
	Store *Store
}

func (r Resolver) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		id := req.Header.Get("X-Project-ID")
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
