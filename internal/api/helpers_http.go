package api

import (
	"encoding/json"
	"infolinks-backend/internal/middleware"
	"net/http"
	"strconv"
	"strings"
)

const maxBodyBytes = 1 << 20

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeJSONError(w http.ResponseWriter, r *http.Request, status int, message string) {
	payload := map[string]string{"error": message}
	if status >= http.StatusInternalServerError && r != nil {
		if id := middleware.RequestIDFromContext(r.Context()); id != "" {
			payload["request_id"] = id
		}
	}
	writeJSON(w, status, payload)
}

func decodeJSONBody(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		writeJSONError(w, r, http.StatusBadRequest, "Invalid request body")
		return false
	}
	return true
}

func parsePaginationParams(r *http.Request, defaultLimit int) (limit int, offset int, q string) {
	limit = defaultLimit
	if limit <= 0 {
		limit = 25
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			if parsed > 100 {
				parsed = 100
			}
			limit = parsed
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}
	q = strings.TrimSpace(r.URL.Query().Get("q"))
	return
}
