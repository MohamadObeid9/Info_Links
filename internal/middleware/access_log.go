package middleware

import (
	"net/http"
	"strings"
)

type accessLogAction int

const (
	accessLogSkip accessLogAction = iota
	accessLogDebug
	accessLogInfo
	accessLogWarn
)

func accessLogDecision(method, path, appEnv string, status int) accessLogAction {
	switch {
	case isNoisyPath(path, method), status == http.StatusNoContent, status == http.StatusNotFound:
		return accessLogSkip
	case status >= 400:
		return accessLogWarn
	case method == http.MethodPost, method == http.MethodPatch, method == http.MethodDelete:
		return accessLogInfo
	case strings.HasPrefix(path, "/api/admin"):
		return accessLogSkip
	default:
		if appEnv == "development" {
			return accessLogInfo
		} else {
			return accessLogDebug
		}
	}
}

func isNoisyPath(path, method string) bool {
	if method == http.MethodHead {
		return true
	}

	switch path {
	case "/metrics", "/healthz", "/readyz", "/robots.txt", "/sitemap.xml", "/.well-known/api-catalog", "/.well-known/oauth-protected-resource", "/.well-known/oauth-authorization-server", "/.well-known/openid-configuration", "/.well-known/jwks.json", "/.well-known/agent-card.json", "/.well-known/agents-index.json", "/.well-known/agent-skills/index.json", "/.well-known/mcp/server-card.json", "/.well-known/http-message-signatures-directory", "/openapi.json", "/auth.md", "/mcp":
		return true
	}
	if strings.HasPrefix(path, "/.well-known/agent-skills/") {
		return true
	}
	if strings.HasPrefix(path, "/.well-known/mcp/") {
		return true
	}

	if strings.HasPrefix(path, "/assets/") {
		return true
	}

	if strings.HasSuffix(path, ".env") || strings.HasSuffix(path, ".php") {
		return true
	}

	if dot := strings.LastIndexByte(path, '.'); dot != -1 && !strings.ContainsRune(path[dot:], '/') {
		return true
	}

	return false
}
