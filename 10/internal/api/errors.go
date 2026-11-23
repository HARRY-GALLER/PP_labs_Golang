package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"userapi/internal/user"
)

// Error — единый формат ошибки API.
type Error struct {
	Status  int               `json:"-"`
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

func (e *Error) Error() string { return e.Code + ": " + e.Message }

func badRequest(code, msg string) *Error {
	return &Error{Status: http.StatusBadRequest, Code: code, Message: msg}
}

func validationFailed(details map[string]string) *Error {
	return &Error{
		Status:  http.StatusUnprocessableEntity,
		Code:    "validation_failed",
		Message: "переданные данные некорректны",
		Details: details,
	}
}

// writeError — единственное место, где ошибки превращаются в HTTP-ответ.
func (a *API) writeError(w http.ResponseWriter, r *http.Request, err error) {
	var apiErr *Error
	switch {
	case errors.As(err, &apiErr):
	case errors.Is(err, user.ErrNotFound):
		apiErr = &Error{Status: http.StatusNotFound, Code: "not_found", Message: "пользователь не найден"}
	case errors.Is(err, user.ErrEmailTaken):
		apiErr = &Error{Status: http.StatusConflict, Code: "email_taken", Message: "пользователь с таким email уже существует"}
	default:
		a.log.Error("internal error", "method", r.Method, "path", r.URL.Path, "err", err)
		apiErr = &Error{Status: http.StatusInternalServerError, Code: "internal_error", Message: "внутренняя ошибка сервера"}
	}
	writeJSON(w, apiErr.Status, map[string]any{"error": apiErr})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// decodeJSON читает тело запроса строго: без неизвестных полей, не больше 1 МБ.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		var tooBig *http.MaxBytesError
		switch {
		case errors.As(err, &tooBig):
			return &Error{Status: http.StatusRequestEntityTooLarge, Code: "body_too_large", Message: "тело запроса слишком большое"}
		case errors.Is(err, io.EOF):
			return badRequest("empty_body", "тело запроса пустое")
		default:
			return badRequest("invalid_json", "некорректный JSON: "+err.Error())
		}
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return badRequest("invalid_json", "после JSON-объекта есть лишние данные")
	}
	return nil
}
