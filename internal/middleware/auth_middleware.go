package middleware

import (
	"errors"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

// RequireAdmin middleware verifies the JWT token in the Authorization header.
func RequireAdmin(jwtSecret string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			writeJSONErr(w, http.StatusUnauthorized, "Unauthorized: No token provided")
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
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			writeJSONErr(w, http.StatusUnauthorized, "Unauthorized: Invalid token")
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			writeJSONErr(w, http.StatusUnauthorized, "Unauthorized: Invalid token claims")
			return
		}
		adminClaim, ok := claims["admin"].(bool)
		if !ok || !adminClaim {
			writeJSONErr(w, http.StatusForbidden, "Forbidden: Admin access required")
			return
		}

		next.ServeHTTP(w, r)
	}
}

func IsAuthenticatedAdmin(jwtSecret, tokenString string) bool {
	if tokenString == "" {
		return false
	}

	// Handle "Bearer <token>" format
	if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
		tokenString = tokenString[7:]
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return false
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return false
	}
	adminClaim, ok := claims["admin"].(bool)
	if !ok || !adminClaim {
		return false
	}
	return true
}

func writeJSONErr(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(`{"error":"` + msg + `"}`))
}
