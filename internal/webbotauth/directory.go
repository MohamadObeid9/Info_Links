package webbotauth

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Directory hosts an Ed25519 key pair for Web Bot Auth (HTTP Message Signatures).
type Directory struct {
	private ed25519.PrivateKey
	public  ed25519.PublicKey
	jwk     map[string]string
	jwks    []byte
	kid     string
	origin  string // e.g. https://example.com — used as Signature-Agent
}

// NewDirectory derives a stable Ed25519 key from seed material (typically JWT_SECRET)
// and prepares the JWKS served at /.well-known/http-message-signatures-directory.
func NewDirectory(seed, siteBaseURL string) (*Directory, error) {
	seed = strings.TrimSpace(seed)
	if seed == "" {
		return nil, fmt.Errorf("web bot auth seed is required")
	}
	sum := sha256.Sum256([]byte("infolinks-web-bot-auth-v1:" + seed))
	priv := ed25519.NewKeyFromSeed(sum[:])
	pub := priv.Public().(ed25519.PublicKey)

	x := base64.RawURLEncoding.EncodeToString(pub)
	kid, err := jwkThumbprint(x)
	if err != nil {
		return nil, err
	}
	jwk := map[string]string{
		"kty": "OKP",
		"crv": "Ed25519",
		"alg": "Ed25519",
		"x":   x,
		"kid": kid,
	}
	jwks, err := json.Marshal(map[string]any{"keys": []any{jwk}})
	if err != nil {
		return nil, err
	}
	origin := strings.TrimSuffix(strings.TrimSpace(siteBaseURL), "/")
	return &Directory{
		private: priv,
		public:  pub,
		jwk:     jwk,
		jwks:    jwks,
		kid:     kid,
		origin:  origin,
	}, nil
}

func jwkThumbprint(x string) (string, error) {
	// RFC 7638 — required members only, lexicographic order.
	payload := `{"crv":"Ed25519","kty":"OKP","x":"` + x + `"}`
	sum := sha256.Sum256([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(sum[:]), nil
}

// JWKS returns the raw key-directory document bytes.
func (d *Directory) JWKS() []byte {
	if d == nil {
		return nil
	}
	out := make([]byte, len(d.jwks))
	copy(out, d.jwks)
	return out
}

// KeyID returns the JWK thumbprint used as Signature-Input keyid.
func (d *Directory) KeyID() string {
	if d == nil {
		return ""
	}
	return d.kid
}

// ServeDirectory writes the signed HTTP Message Signatures directory response.
func (d *Directory) ServeDirectory(w http.ResponseWriter, r *http.Request) {
	if d == nil {
		http.Error(w, "web bot auth not configured", http.StatusServiceUnavailable)
		return
	}
	body := d.JWKS()
	authority := requestAuthority(r)
	created := time.Now().UTC().Unix()
	expires := created + 300

	contentDigest := "sha-256=:" + base64.StdEncoding.EncodeToString(sha256Sum(body)) + ":"
	params := fmt.Sprintf(
		`("@authority";req "content-digest");created=%d;expires=%d;alg="ed25519";keyid="%s";tag="http-message-signatures-directory"`,
		created, expires, d.kid,
	)
	base := signatureBase(map[string]string{
		`"@authority";req`: authority,
		`"content-digest"`: contentDigest,
	}, []string{`"@authority";req`, `"content-digest"`}, params)
	sig := ed25519.Sign(d.private, []byte(base))

	w.Header().Set("Content-Type", "application/http-message-signatures-directory+json")
	w.Header().Set("Content-Digest", contentDigest)
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Header().Set("Signature-Input", "sig1="+params)
	w.Header().Set("Signature", "sig1=:"+base64.StdEncoding.EncodeToString(sig)+":")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}

// SignRequest attaches Web Bot Auth headers to an outbound request.
func (d *Directory) SignRequest(req *http.Request) error {
	if d == nil {
		return fmt.Errorf("web bot auth directory is nil")
	}
	if req == nil {
		return fmt.Errorf("request is nil")
	}
	created := time.Now().UTC().Unix()
	expires := created + 60
	agent := d.origin
	if agent == "" {
		agent = "https://localhost"
	}
	authority := req.URL.Host
	if authority == "" && req.Host != "" {
		authority = req.Host
	}

	req.Header.Set("Signature-Agent", strconv.Quote(agent))
	params := fmt.Sprintf(
		`("@authority" "signature-agent");created=%d;expires=%d;keyid="%s";alg="ed25519";tag="web-bot-auth"`,
		created, expires, d.kid,
	)
	base := signatureBase(map[string]string{
		`"@authority"`:      authority,
		`"signature-agent"`: strconv.Quote(agent),
	}, []string{`"@authority"`, `"signature-agent"`}, params)
	sig := ed25519.Sign(d.private, []byte(base))
	req.Header.Set("Signature-Input", "sig1="+params)
	req.Header.Set("Signature", "sig1=:"+base64.StdEncoding.EncodeToString(sig)+":")
	return nil
}

func signatureBase(values map[string]string, order []string, params string) string {
	var b strings.Builder
	for _, comp := range order {
		b.WriteString(comp)
		b.WriteString(": ")
		b.WriteString(values[comp])
		b.WriteByte('\n')
	}
	b.WriteString(`"@signature-params": `)
	b.WriteString(params)
	return b.String()
}

func requestAuthority(r *http.Request) string {
	if r == nil {
		return ""
	}
	if host := strings.TrimSpace(r.Host); host != "" {
		return host
	}
	if r.URL != nil && r.URL.Host != "" {
		return r.URL.Host
	}
	return ""
}

func sha256Sum(b []byte) []byte {
	sum := sha256.Sum256(b)
	return sum[:]
}
