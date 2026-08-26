package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/golang-jwt/jwt/v5"
)

type userContextKey struct{}

// authenticatedUser is the student identity carried by a student JWT.
type authenticatedUser struct {
	id      int
	isGuest bool
}

var (
	errNoToken       = errors.New("no token provided")
	errInvalidToken  = errors.New("invalid token")
	errInvalidClaims = errors.New("invalid token claims")

	resourceMetadataMu  sync.RWMutex
	resourceMetadataURL = "/.well-known/oauth-protected-resource"
)

// SetResourceMetadataURL sets the absolute (preferred) or path-only URI used in
// WWW-Authenticate resource_metadata on 401 responses (RFC 9728).
func SetResourceMetadataURL(url string) {
	url = strings.TrimSpace(url)
	if url == "" {
		url = "/.well-known/oauth-protected-resource"
	}
	resourceMetadataMu.Lock()
	resourceMetadataURL = url
	resourceMetadataMu.Unlock()
}

func resourceMetadataParam() string {
	resourceMetadataMu.RLock()
	defer resourceMetadataMu.RUnlock()
	return resourceMetadataURL
}

// RequireAdmin middleware verifies the JWT token in the Authorization header.
func RequireAdmin(jwtSecret string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, err := parseHS256Claims(jwtSecret, r.Header.Get("Authorization"))
		if err != nil {
			writeAuthErr(w, err)
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

// RequireUser accepts any student token, guests included, and puts the student
// identity in the request context.
func RequireUser(jwtSecret string, next http.HandlerFunc) http.HandlerFunc {
	return requireUser(jwtSecret, true, next)
}

// RequireRegisteredUser rejects guests, for actions gated behind signup.
func RequireRegisteredUser(jwtSecret string, next http.HandlerFunc) http.HandlerFunc {
	return requireUser(jwtSecret, false, next)
}

func requireUser(jwtSecret string, allowGuest bool, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, err := parseHS256Claims(jwtSecret, r.Header.Get("Authorization"))
		if err != nil {
			writeAuthErr(w, err)
			return
		}
		user, ok := userFromClaims(claims)
		if !ok {
			writeJSONErr(w, http.StatusUnauthorized, "Unauthorized: Invalid token claims")
			return
		}
		if user.isGuest && !allowGuest {
			writeJSONErr(w, http.StatusForbidden, "Forbidden: Registered account required")
			return
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userContextKey{}, user)))
	}
}

// UserIDFromContext returns the student id stored by RequireUser, or 0.
func UserIDFromContext(ctx context.Context) int {
	if user, ok := ctx.Value(userContextKey{}).(authenticatedUser); ok {
		return user.id
	}
	return 0
}

// UserIsGuestFromContext reports whether the authenticated student is still a guest.
func UserIsGuestFromContext(ctx context.Context) bool {
	if user, ok := ctx.Value(userContextKey{}).(authenticatedUser); ok {
		return user.isGuest
	}
	return false
}

// ContextWithUser stores a student identity in ctx the way RequireUser does.
func ContextWithUser(ctx context.Context, userID int, isGuest bool) context.Context {
	return context.WithValue(ctx, userContextKey{}, authenticatedUser{id: userID, isGuest: isGuest})
}

// GuestIDFromHeader reads the guest id from an optional student token. It returns
// 0 when the header is absent, expired, invalid or already belongs to a
// registered student, so callers can treat the guest session as best effort.
func GuestIDFromHeader(jwtSecret, authHeader string) int {
	claims, err := parseHS256Claims(jwtSecret, authHeader)
	if err != nil {
		return 0
	}
	user, ok := userFromClaims(claims)
	if !ok || !user.isGuest {
		return 0
	}
	return user.id
}

func IsAuthenticatedAdmin(jwtSecret, tokenString string) bool {
	claims, err := parseHS256Claims(jwtSecret, tokenString)
	if err != nil {
		return false
	}
	adminClaim, ok := claims["admin"].(bool)
	return ok && adminClaim
}

// parseHS256Claims validates an HS256 token from an Authorization header value,
// with or without the "Bearer " prefix.
func parseHS256Claims(jwtSecret, authHeader string) (jwt.MapClaims, error) {
	tokenString := strings.TrimSpace(authHeader)
	if tokenString == "" {
		return nil, errNoToken
	}
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, errInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errInvalidClaims
	}
	return claims, nil
}

// userFromClaims reads the student claims. JSON numbers decode as float64.
func userFromClaims(claims jwt.MapClaims) (authenticatedUser, bool) {
	rawID, ok := claims["userID"].(float64)
	if !ok || rawID < 1 {
		return authenticatedUser{}, false
	}
	isGuest, _ := claims["isGuest"].(bool)
	return authenticatedUser{id: int(rawID), isGuest: isGuest}, true
}

func writeAuthErr(w http.ResponseWriter, err error) {
	w.Header().Set("WWW-Authenticate", fmt.Sprintf(
		`Bearer FAKESECRET_g3h4i5j6k7l8m9n0o1p2="%s"`,
		resourceMetadataParam(),
	))
	switch {
	case errors.Is(err, errNoToken):
		writeJSONErr(w, http.StatusUnauthorized, "Unauthorized: No token provided")
	case errors.Is(err, errInvalidClaims):
		writeJSONErr(w, http.StatusUnauthorized, "Unauthorized: Invalid token claims")
	default:
		writeJSONErr(w, http.StatusUnauthorized, "Unauthorized: Invalid token")
	}
}

func writeJSONErr(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(`{"error":"` + msg + `"}`))
}
