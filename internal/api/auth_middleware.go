package api

import (
	"errors"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

// RequireAdmin middleware verifies the JWT token in the Authorization header.
func (h *Handler) requireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			writeJSONError(w, http.StatusUnauthorized, "Unauthorized: No token provided")
			return
		}

		// Handle "Bearer <token>" format
		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return h.jwtSecret, nil
		})

		if err != nil || !token.Valid {
			writeJSONError(w, http.StatusUnauthorized, "Unauthorized: Invalid token")
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "Unauthorized: Invalid token claims")
			return
		}
		adminClaim, ok := claims["admin"].(bool)
		if !ok || !adminClaim {
			writeJSONError(w, http.StatusForbidden, "Forbidden: Admin access required")
			return
		}

		next.ServeHTTP(w, r)
	}
}
