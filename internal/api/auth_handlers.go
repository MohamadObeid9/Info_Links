package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// HandleLogin authenticates an admin and returns a JWT.
func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if !decodeJSONBody(w, r, &creds) {
		return
	}

	supabaseToken, err := h.verifyWithSupabase(r, creds.Email, creds.Password)
	if err != nil {
		writeJSONError(w, r, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	if !isAdmin(supabaseToken) {
		writeJSONError(w, r, http.StatusForbidden, "Forbidden: Admin access required")
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"admin": true,
		"exp":   time.Now().Add(time.Hour * 24 * 7).Unix(), // 1 week
	})

	tokenString, err := token.SignedString(h.jwtSecret)
	if err != nil {
		writeJSONError(w, r, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"token": tokenString})
}

func (h *Handler) verifyWithSupabase(r *http.Request, email, password string) (string, error) {
	payload := map[string]string{"email": email, "password": password}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("json marshal: %w", err)
	}
	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, h.supabaseURL+"/auth/v1/token?grant_type=password", bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", h.supbaseAnonKey)

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("response status code: %d", resp.StatusCode)
	}

	body, err = io.ReadAll(io.LimitReader(resp.Body, 8192))
	if err != nil {
		return "", err
	}

	var result struct {
		AccessToken string `json:"access_token"`
	}

	if err := json.Unmarshal(body, &result); err != nil || result.AccessToken == "" {
		return "", fmt.Errorf("supabase auth: no access_token in response")
	}

	return result.AccessToken, nil
}

func isAdmin(supabaseToken string) bool {
	token, _, err := jwt.NewParser().ParseUnverified(supabaseToken, jwt.MapClaims{})
	if err != nil {
		return false
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return false
	}

	meta, ok := claims["app_metadata"].(map[string]any)
	if !ok {
		return false
	}

	role, _ := meta["role"].(string)
	return role == "admin"
}
