package api

import (
	"crypto/subtle"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"userapi/internal/auth"
	"userapi/internal/user"
)

// authorize проверяет токен каждого запроса к защищенным маршрутам.
func (a *API) authorize(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := bearerToken(r)
		if token == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]any{
				"error": &Error{Code: "unauthorized", Message: "требуется заголовок Authorization: Bearer <token>"},
			})
			return
		}
		session, ok := a.auth.Check(token)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]any{
				"error": &Error{Code: "invalid_token", Message: "токен неверный или срок его действия истек"},
			})
			return
		}
		// Для изменения состояния требуется второй, независимый CSRF-токен сессии.
		if r.Method != http.MethodGet && r.Method != http.MethodHead && r.Method != http.MethodOptions {
			provided := r.Header.Get("X-CSRF-Token")
			if len(provided) != len(session.CSRFToken) || subtle.ConstantTimeCompare([]byte(provided), []byte(session.CSRFToken)) != 1 {
				writeJSON(w, http.StatusForbidden, map[string]any{"error": &Error{Code: "invalid_csrf", Message: "неверный или отсутствует X-CSRF-Token"}})
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// Права администратора нужны только для изменения записей пользователей.
func (a *API) adminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, ok := a.auth.Check(bearerToken(r))
		if !ok || session.Role != "admin" {
			writeJSON(w, http.StatusForbidden, map[string]any{"error": &Error{Code: "forbidden", Message: "доступ разрешён только администратору"}})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func bearerToken(r *http.Request) string {
	v := r.Header.Get("Authorization")
	if !strings.HasPrefix(v, "Bearer ") {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(v, "Bearer "))
}

func (a *API) register(w http.ResponseWriter, r *http.Request) error {
	var in user.Input
	if err := decodeJSON(w, r, &in); err != nil {
		return err
	}
	in.Normalize()
	fields := in.Validate()
	if err := auth.ValidatePassword(in.Password); err != nil {
		fields["password"] = err.Error()
	}
	if len(fields) > 0 {
		return validationFailed(fields)
	}
	u, err := a.auth.Register(r.Context(), in)
	if err != nil {
		return err
	}
	w.Header().Set("Location", "/users/"+strconv.FormatInt(u.ID, 10))
	writeJSON(w, http.StatusCreated, u)
	return nil
}

func (a *API) login(w http.ResponseWriter, r *http.Request) error {
	var in struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := decodeJSON(w, r, &in); err != nil {
		return err
	}
	token, sess, err := a.auth.Login(r.Context(), in.Email, in.Password)
	if errors.Is(err, auth.ErrInvalidCredentials) {
		return &Error{Status: http.StatusUnauthorized, Code: "invalid_credentials", Message: "неверный email или пароль"}
	}
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"token":      token,
		"expires_at": sess.ExpiresAt,
		"user_id":    sess.UserID,
		"role":       sess.Role,
		"csrf_token": sess.CSRFToken,
	})
	return nil
}

func (a *API) logout(w http.ResponseWriter, r *http.Request) error {
	a.auth.Logout(bearerToken(r))
	w.WriteHeader(http.StatusNoContent)
	return nil
}
