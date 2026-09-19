package licx

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

const (
	wireVersion byte = 1
	chunkSize   int  = 5
)

var (
	ErrInvalidFormat  = errors.New("license format is invalid")
	ErrInvalidVersion = errors.New("invalid license version")
	ErrInvalidSig     = errors.New("signature verification failed")
	ErrExpired        = errors.New("license has expired")
	ErrAppMismatch    = errors.New("license issued for a different application")
)

var pubkey string

type Claims struct {
	Version   int    `json:"v"`
	App       string `json:"app"` // Must equal the app passed to Issue/Verify (case-sensitive)
	Plan      string `json:"plan"`
	Email     string `json:"email"`
	ExpiresAt int64  `json:"exp"` // Unix timestamp, 0 = perpetual
}

func (c *Claims) IsExpired() bool {
	if c.ExpiresAt == 0 {
		return false
	}
	return time.Now().UTC().Unix() > c.ExpiresAt
}

// ParsePrivateKey decodes the base64 64-byte Ed25519 private key
// (seed + public key) produced by GenerateKey, the format the issuing
// side stores.
func ParsePrivateKey(privateB64 string) (ed25519.PrivateKey, error) {
	raw, err := base64.StdEncoding.DecodeString(privateB64)
	if err != nil || len(raw) != ed25519.PrivateKeySize {
		return nil, errors.New("private key must be the base64 64-byte Ed25519 key from GenerateKey")
	}
	return ed25519.PrivateKey(raw), nil
}

// Issue generates a formatted key using the worker's private key.
//
// app names the license's owner app. With a shared keypair across apps,
// binding is enforced by the signed Claims.App value, which must equal
// app (case-sensitive). The uppercase app name also becomes the readable
// key prefix, e.g. Issue("myapp", ...) emits "MYAPP-XXXXX-...".
func Issue(app string, claims Claims, priv ed25519.PrivateKey) (string, error) {
	if app == "" {
		return "", errors.New("app name is required")
	}
	claims.App = app
	claims.Version = int(wireVersion)

	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	sig := ed25519.Sign(priv, payload)

	// [wireVersion (1B)] + [payload (NB)] + [sig (64B)]
	data := make([]byte, 0, 1+len(payload)+len(sig))
	data = append(data, wireVersion)
	data = append(data, payload...)
	data = append(data, sig...)

	encoded := base64.RawStdEncoding.EncodeToString(data)
	return strings.ToUpper(app) + "-" + chunkKey(encoded), nil
}

func chunkKey(encoded string) string {
	var b strings.Builder
	b.Grow(len(encoded) + len(encoded)/chunkSize)
	for i := 0; i < len(encoded); i += chunkSize {
		if i > 0 {
			b.WriteByte('-')
		}
		end := min(i+chunkSize, len(encoded))
		b.WriteString(encoded[i:end])
	}
	return b.String()
}

// Verify extracts and validates claims against a public key.
//
// app must match the signed Claims.App exactly (case-sensitive), and the
// key must carry the matching uppercase prefix, e.g. a key issued with
// Issue("myapp", ...) verifies with Verify(key, "myapp").
//
// An expired key returns the parsed claims alongside ErrExpired, so
// callers can show plan or email details in expiry messages.
//
// The public key is injected at build time via ldflags:
//
//	-X github.com/mizuchilabs/kata/licx.pubkey=<hex>
//
// or at runtime with SetPublicKey.
func Verify(rawKey, app string) (*Claims, error) {
	prefix := strings.ToUpper(app) + "-"

	enc, ok := strings.CutPrefix(strings.TrimSpace(rawKey), prefix)
	if !ok {
		return nil, ErrInvalidFormat
	}

	clean := strings.ReplaceAll(enc, "-", "")
	raw, err := base64.RawStdEncoding.DecodeString(clean)
	if err != nil || len(raw) < 1+ed25519.SignatureSize+1 {
		return nil, ErrInvalidFormat
	}

	if raw[0] != wireVersion {
		return nil, ErrInvalidFormat
	}

	sigLen := ed25519.SignatureSize
	payload := raw[1 : len(raw)-sigLen]
	sig := raw[len(raw)-sigLen:]
	pub, err := loadLicensePubKey()
	if err != nil {
		return nil, err
	}

	if !ed25519.Verify(pub, payload, sig) {
		return nil, ErrInvalidSig
	}

	var c Claims
	if err := json.Unmarshal(payload, &c); err != nil {
		return nil, ErrInvalidFormat
	}

	if c.App != app {
		return nil, ErrAppMismatch
	}
	if c.Version != int(wireVersion) {
		return nil, ErrInvalidVersion
	}
	if c.IsExpired() {
		return &c, ErrExpired
	}

	return &c, nil
}

// SetPublicKey sets the verification key at runtime, for tests and
// embedders that do not use ldflags. The key is hex-encoded, the same
// format the ldflags-injected pubkey var expects. An empty string clears
// the key.
func SetPublicKey(pubHex string) error {
	if strings.TrimSpace(pubHex) == "" {
		pubkey = ""
		return nil
	}
	b, err := hex.DecodeString(strings.TrimSpace(pubHex))
	if err != nil || len(b) != ed25519.PublicKeySize {
		return errors.New("invalid ed25519 public key hex")
	}
	pubkey = strings.TrimSpace(pubHex)
	return nil
}

// HasPublicKey reports whether a verification key is available, via
// ldflags injection or SetPublicKey. Verifiers that surface a distinct
// "not configured" state use it before Verify.
func HasPublicKey() bool {
	return strings.TrimSpace(pubkey) != ""
}

// loadLicensePubKey decodes the hex-encoded ldflags-injected public key.
func loadLicensePubKey() (ed25519.PublicKey, error) {
	if pubkey == "" {
		return nil, errors.New("no license public key injected")
	}
	b, err := hex.DecodeString(strings.TrimSpace(pubkey))
	if err != nil || len(b) != ed25519.PublicKeySize {
		return nil, errors.New("invalid ed25519 public key hex")
	}
	return ed25519.PublicKey(b), nil
}
