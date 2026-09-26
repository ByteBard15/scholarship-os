package response

import (
	"encoding/json"
	"net/http"
)

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func JSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func Data(w http.ResponseWriter, status int, data any) { JSON(w, status, map[string]any{"data": data}) }
func Collection(w http.ResponseWriter, data any, count int) {
	JSON(w, http.StatusOK, map[string]any{"data": data, "meta": map[string]int{"count": count}})
}
func Error(w http.ResponseWriter, status int, code, message string) {
	JSON(w, status, map[string]any{"error": ErrorBody{Code: code, Message: message}})
}
