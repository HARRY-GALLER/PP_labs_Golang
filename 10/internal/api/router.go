// Package api содержит HTTP-слой: маршруты, обработчики, middleware.
package api

import (
	"log/slog"
	"net/http"
	"time"

	"userapi/internal/auth"
	"userapi/internal/user"
)

type API struct {
	repo user.Repository
	log  *slog.Logger
	auth *auth.Manager
}

type handlerFunc func(w http.ResponseWriter, r *http.Request) error

// handle превращает обработчик, возвращающий error, в обычный http.HandlerFunc.
func (a *API) handle(h handlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			a.writeError(w, r, err)
		}
	}
}

func NewRouter(repo user.Repository, manager *auth.Manager, log *slog.Logger) http.Handler {
	a := &API{repo: repo, auth: manager, log: log}

	mux := http.NewServeMux() // шаблоны «METHOD /path/{id}» требуют Go 1.22+
	mux.HandleFunc("POST /auth/register", a.handle(a.register))
	mux.HandleFunc("POST /auth/login", a.handle(a.login))
	mux.Handle("POST /auth/logout", a.authorize(a.handle(a.logout)))
	mux.Handle("GET /users", a.authorize(a.handle(a.listUsers)))
	mux.Handle("GET /users/{id}", a.authorize(a.handle(a.getUser)))
	mux.Handle("POST /users", a.authorize(a.adminOnly(a.handle(a.createUser))))
	mux.Handle("PUT /users/{id}", a.authorize(a.adminOnly(a.handle(a.updateUser))))
	mux.Handle("DELETE /users/{id}", a.authorize(a.adminOnly(a.handle(a.deleteUser))))

	return a.recoverer(a.logging(mux))
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

func (a *API) logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		a.log.Info("request", "method", r.Method, "path", r.URL.Path,
			"status", rec.status, "duration", time.Since(start))
	})
}

func (a *API) recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				a.log.Error("panic", "err", rec)
				writeJSON(w, http.StatusInternalServerError, map[string]any{
					"error": &Error{Code: "internal_error", Message: "внутренняя ошибка сервера"},
				})
			}
		}()
		next.ServeHTTP(w, r)
	})
}
