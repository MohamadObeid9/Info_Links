package api

import "net/http"

func (h *Handler) handleHTTPMessageSignaturesDirectory(w http.ResponseWriter, r *http.Request) {
	if h.webBotAuth == nil {
		writeJSONError(w, r, http.StatusServiceUnavailable, "Web Bot Auth is not configured")
		return
	}
	h.webBotAuth.ServeDirectory(w, r)
}
