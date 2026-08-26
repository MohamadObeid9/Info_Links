package api

import (
	"encoding/json"
	"net/http"
	"strings"
)

// handleAuthMD serves /auth.md — prose companion for agent registration discovery.
func (h *Handler) handleAuthMD(w http.ResponseWriter, r *http.Request) {
	base := h.baseURL()
	var b strings.Builder
	b.WriteString("# auth.md\n\n")
	b.WriteString("You are an agent registering for **Info Links**, a CNAM Liban student materials hub.\n\n")
	b.WriteString("- **Resource server (API):** `")
	b.WriteString(base)
	b.WriteString("`\n")
	b.WriteString("- **Authorization server:** `")
	b.WriteString(base)
	b.WriteString("` (same origin)\n")
	b.WriteString("- **Audience:** automated clients and agents that need a **student** session JWT to call public analytics and submission endpoints. Admin APIs are out of scope for agent registration.\n\n")

	b.WriteString("## 1. Discover\n\n")
	b.WriteString("1. On `401`, read `WWW-Authenticate: Bearer resource_metadata=\"…\"` if present, or fetch Protected Resource Metadata at:\n\n")
	b.WriteString("```http\nGET ")
	b.WriteString(base)
	b.WriteString("/.well-known/oauth-protected-resource\nAccept: application/json\n```\n\n")
	b.WriteString("2. From PRM, note `resource`, `authorization_servers`, `scopes_supported`, and `bearer_methods_supported` (`header`).\n\n")
	b.WriteString("3. Fetch Authorization Server / OIDC discovery (issuer matches PRM):\n\n")
	b.WriteString("```http\nGET ")
	b.WriteString(base)
	b.WriteString("/.well-known/openid-configuration\nAccept: application/json\n```\n\n")
	b.WriteString("Same document is also at `/.well-known/oauth-authorization-server`. It includes `jwks_uri` (HS256 shared-secret tokens — JWKS `keys` is empty).\n\n")
	b.WriteString("4. Read `agent_auth`: `skill` (this file), `register_uri`, `identity_types_supported`, and the `anonymous` method (`claim_uri`, credential types).\n\n")

	b.WriteString("## 2. Pick a method\n\n")
	b.WriteString("Info Links supports **anonymous** registration only (no email, no ID-JAG).\n\n")
	b.WriteString("| You have | Method |\n|---|---|\n")
	b.WriteString("| Nothing yet | `anonymous` → `POST` `register_uri` (`/api/users/guest`) |\n")
	b.WriteString("| Guest bearer + name/number | Claim → `POST` `claim_uri` (`/api/users/register`) |\n")
	b.WriteString("| Existing name + number | Sign-in → `POST /api/users/login` |\n\n")

	b.WriteString("## 3. Register (anonymous)\n\n")
	b.WriteString("Mint a guest student JWT (no body):\n\n")
	b.WriteString("```http\nPOST ")
	b.WriteString(base)
	b.WriteString("/api/users/guest\nAccept: application/json\n```\n\n")
	b.WriteString("```json\n{\n  \"token\": \"<guest_jwt>\"\n}\n```\n\n")
	b.WriteString("Use `Authorization: Bearer <guest_jwt>` for browse analytics (`POST /api/page_views`, search, browse). Guests **cannot** open gated links, submit reports, or favorite courses.\n\n")

	b.WriteString("## 4. Claim (upgrade guest → registered student)\n\n")
	b.WriteString("Send the guest bearer and a unique first name + last name + number (1–100):\n\n")
	b.WriteString("```http\nPOST ")
	b.WriteString(base)
	b.WriteString("/api/users/register\nAuthorization: Bearer <guest_jwt>\nContent-Type: application/json\n\n")
	b.WriteString("{\n  \"first_name\": \"ziad\",\n  \"last_name\": \"baroudi\",\n  \"number\": 65\n}\n```\n\n")
	b.WriteString("```json\n{\n  \"token\": \"<student_jwt>\",\n  \"user\": { \"id\": 1, \"is_guest\": false, \"handle\": \"ziad_baroudi_65\" }\n}\n```\n\n")
	b.WriteString("On `409`, pick another `number`. The claimed row keeps the same user id as the guest (pre-claim activity stays attached).\n\n")

	b.WriteString("## 5. Use the access token\n\n")
	b.WriteString("Present the JWT as a Bearer header (`bearer_methods_supported: header`):\n\n")
	b.WriteString("```http\nGET ")
	b.WriteString(base)
	b.WriteString("/api/users/me\nAuthorization: Bearer <student_jwt>\n```\n\n")
	b.WriteString("Registered-student scopes unlock `POST /api/link_clicks`, reports, feedback, contributions, and favorites. See OpenAPI at `/openapi.json`.\n\n")

	b.WriteString("## 6. Errors\n\n")
	b.WriteString("| Status | Meaning | Agent action |\n|---|---|---|\n")
	b.WriteString("| 401 | Missing/invalid JWT | Re-register at `register_uri` |\n")
	b.WriteString("| 403 | Guest hitting registered-only route | Claim via `claim_uri` |\n")
	b.WriteString("| 409 | Name+number taken | Retry register with another number |\n")
	b.WriteString("| 404 on login | Unknown credentials | Use register instead |\n\n")

	b.WriteString("## Notes\n\n")
	b.WriteString("- There is no OAuth refresh_token; keep using the issued JWT until it fails, then obtain a new one.\n")
	b.WriteString("- Admin login (`POST /api/auth/login`) is human/operator only — not part of agent registration.\n")
	b.WriteString("- If this file conflicts with `/.well-known/oauth-protected-resource`, trust the PRM.\n")

	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(b.String()))
}

// handleOAuthProtectedResource serves RFC 9728 Protected Resource Metadata.
func (h *Handler) handleOAuthProtectedResource(w http.ResponseWriter, r *http.Request) {
	base := h.baseURL()
	doc := map[string]any{
		"resource":                 base + "/",
		"resource_name":            "Info Links API",
		"authorization_servers":    []string{base},
		"scopes_supported":         []string{"student", "student:registered"},
		"bearer_methods_supported": []string{"header"},
		"resource_documentation":   h.absURL("/auth.md"),
	}
	writeDiscoveryJSON(w, doc)
}

// oauthAuthorizationServerMetadata is shared by RFC 8414 and OIDC discovery.
func (h *Handler) oauthAuthorizationServerMetadata() map[string]any {
	base := h.baseURL()
	return map[string]any{
		"issuer":                                base,
		"authorization_endpoint":                base + "/",
		"token_endpoint":                        h.absURL("/api/users/guest"),
		"registration_endpoint":                 h.absURL("/api/users/register"),
		"jwks_uri":                              h.absURL("/.well-known/jwks.json"),
		"response_types_supported":              []string{"token"},
		"grant_types_supported":                 []string{"urn:infolinks:grant-type:anonymous-guest"},
		"token_endpoint_auth_methods_supported": []string{"none"},
		"scopes_supported":                      []string{"student", "student:registered"},
		"id_token_signing_alg_values_supported": []string{"HS256"},
		"subject_types_supported":               []string{"public"},
		"agent_auth": map[string]any{
			"skill":                    h.absURL("/auth.md"),
			"register_uri":             h.absURL("/api/users/guest"),
			"identity_types_supported": []string{"anonymous"},
			"anonymous": map[string]any{
				"credential_types_supported": []string{"access_token"},
				"claim_uri":                  h.absURL("/api/users/register"),
			},
		},
	}
}

// handleOAuthAuthorizationServer serves RFC 8414 Authorization Server metadata
// plus an agent_auth discovery block for anonymous student registration.
func (h *Handler) handleOAuthAuthorizationServer(w http.ResponseWriter, r *http.Request) {
	writeDiscoveryJSON(w, h.oauthAuthorizationServerMetadata())
}

// handleOpenIDConfiguration serves OpenID Connect Discovery 1.0 metadata
// (same core fields as RFC 8414, including jwks_uri).
func (h *Handler) handleOpenIDConfiguration(w http.ResponseWriter, r *http.Request) {
	writeDiscoveryJSON(w, h.oauthAuthorizationServerMetadata())
}

// handleJWKS serves a JWKS document. Student/admin JWTs are HS256 (shared secret),
// so there are no public verification keys to publish — keys is an empty set.
func (h *Handler) handleJWKS(w http.ResponseWriter, r *http.Request) {
	writeDiscoveryJSON(w, map[string]any{"keys": []any{}})
}

func writeDiscoveryJSON(w http.ResponseWriter, doc map[string]any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(doc)
}
