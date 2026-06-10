package api

import (
	"net/http"
	"os"
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

	// For simplicity, we compare against an environment variable.
	// In a full multi-user system, we would query the users table.
	adminEmail := os.Getenv("ADMIN_EMAIL")
	adminPass := os.Getenv("ADMIN_PASSWORD")

	if creds.Email != adminEmail || creds.Password != adminPass {
		writeJSONError(w, r, http.StatusUnauthorized, "Invalid credentials")
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
