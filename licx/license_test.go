package licx

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"testing"
	"time"
)

func testKeypair(t *testing.T) (ed25519.PublicKey, ed25519.PrivateKey) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	pubkey = hex.EncodeToString(pub)
	t.Cleanup(func() { pubkey = "" })
	return pub, priv
}

func TestIssueVerifyRoundtrip(t *testing.T) {
	_, priv := testKeypair(t)

	key, err := Issue("myapp", Claims{Plan: "pro", Email: "a@b.c"}, priv)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(key, "MYAPP-") {
		t.Fatalf("prefix missing: %s", key)
	}

	c, err := Verify(key, "myapp")
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if c.App != "myapp" || c.Plan != "pro" || c.Email != "a@b.c" || c.Version != int(wireVersion) {
		t.Fatalf("unexpected claims: %+v", c)
	}
}

func TestIssueStampsAppAndVersion(t *testing.T) {
	_, priv := testKeypair(t)

	key, err := Issue("myapp", Claims{App: "otherapp", Plan: "pro"}, priv)
	if err != nil {
		t.Fatal(err)
	}
	// Issue must overwrite caller-supplied App so the key verifies.
	if _, err := Verify(key, "myapp"); err != nil {
		t.Fatalf("stamped App ignored: %v", err)
	}
}

func TestIssueEmptyApp(t *testing.T) {
	_, priv := testKeypair(t)
	if _, err := Issue("", Claims{Plan: "pro"}, priv); err == nil {
		t.Fatal("empty app should error")
	}
}

func TestVerifyAppBinding(t *testing.T) {
	_, priv := testKeypair(t)

	key, err := Issue("myapp", Claims{Plan: "pro"}, priv)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := Verify(key, "otherapp"); !errors.Is(err, ErrInvalidFormat) {
		t.Fatalf("different app prefix: got %v", err)
	}
	if _, err := Verify(key, "MYAPP"); !errors.Is(err, ErrAppMismatch) {
		t.Fatalf("casing must matter: got %v", err)
	}
}

func TestVerifyExpiry(t *testing.T) {
	_, priv := testKeypair(t)

	past := time.Now().UTC().Add(-time.Hour).Unix()
	key, err := Issue("myapp", Claims{Plan: "pro", ExpiresAt: past}, priv)
	if err != nil {
		t.Fatal(err)
	}
	c, err := Verify(key, "myapp")
	if !errors.Is(err, ErrExpired) {
		t.Fatalf("expired key: got %v", err)
	}
	if c == nil || c.Plan != "pro" {
		t.Fatalf("expired key must return claims: %+v", c)
	}

	future := time.Now().UTC().Add(time.Hour).Unix()
	key, err = Issue("myapp", Claims{Plan: "pro", ExpiresAt: future}, priv)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Verify(key, "myapp"); err != nil {
		t.Fatalf("unexpired key: %v", err)
	}
}

func TestParsePrivateKey(t *testing.T) {
	privB64, _, err := GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	priv, err := ParsePrivateKey(privB64)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	pubkey = hex.EncodeToString([]byte(priv.Public().(ed25519.PublicKey)))
	defer func() { pubkey = "" }()

	key, err := Issue("myapp", Claims{Plan: "pro"}, priv)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Verify(key, "myapp"); err != nil {
		t.Fatalf("parsed key must sign valid keys: %v", err)
	}

	if _, err := ParsePrivateKey("not-base64!!"); err == nil {
		t.Fatal("garbage should error")
	}
	if _, err := ParsePrivateKey(base64.StdEncoding.EncodeToString(make([]byte, 16))); err == nil {
		t.Fatal("wrong length should error")
	}
}

func TestSetPublicKey(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	old := pubkey
	defer func() { pubkey = old }()

	pubkey = ""
	if err := SetPublicKey("zzzz"); err == nil {
		t.Fatal("malformed hex should error")
	}
	if err := SetPublicKey(hex.EncodeToString(pub)); err != nil {
		t.Fatalf("valid key rejected: %v", err)
	}

	key, err := Issue("myapp", Claims{Plan: "pro"}, priv)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Verify(key, "myapp"); err != nil {
		t.Fatalf("runtime-set key must verify: %v", err)
	}
}

func TestVerifyTampered(t *testing.T) {
	_, priv := testKeypair(t)

	key, err := Issue("myapp", Claims{Plan: "pro"}, priv)
	if err != nil {
		t.Fatal(err)
	}
	enc, _ := strings.CutPrefix(key, "MYAPP-")
	clean := strings.ReplaceAll(enc, "-", "")
	raw, err := base64.RawStdEncoding.DecodeString(clean)
	if err != nil {
		t.Fatal(err)
	}
	raw[len(raw)-1] ^= 0x01

	tampered := "MYAPP-" + chunkKey(base64.RawStdEncoding.EncodeToString(raw))
	if _, err := Verify(tampered, "myapp"); !errors.Is(err, ErrInvalidSig) {
		t.Fatalf("tampered key: got %v", err)
	}
}

func TestHasPublicKey(t *testing.T) {
	old := pubkey
	defer func() { pubkey = old }()

	pubkey = ""
	if HasPublicKey() {
		t.Fatal("empty key should report false")
	}
	pubkey = "ab12"
	if !HasPublicKey() {
		t.Fatal("set key should report true")
	}
}

func TestVerifyNoPubKeyInjected(t *testing.T) {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	pubkey = ""
	defer func() { pubkey = "" }()

	key, err := Issue("myapp", Claims{Plan: "pro"}, priv)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Verify(key, "myapp"); err == nil {
		t.Fatal("missing public key should error, not panic")
	}
}

func TestVerifyWrongWireVersion(t *testing.T) {
	_, priv := testKeypair(t)

	key, err := Issue("myapp", Claims{Plan: "pro"}, priv)
	if err != nil {
		t.Fatal(err)
	}
	enc, _ := strings.CutPrefix(key, "MYAPP-")
	clean := strings.ReplaceAll(enc, "-", "")
	raw, err := base64.RawStdEncoding.DecodeString(clean)
	if err != nil {
		t.Fatal(err)
	}
	raw[0] = wireVersion + 1

	bumped := "MYAPP-" + chunkKey(base64.RawStdEncoding.EncodeToString(raw))
	if _, err := Verify(bumped, "myapp"); !errors.Is(err, ErrInvalidFormat) {
		t.Fatalf("wrong wire version: got %v", err)
	}
}

func TestGenerateKey(t *testing.T) {
	privB64, pubHex, err := GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	priv, err := base64.StdEncoding.DecodeString(privB64)
	if err != nil || len(priv) != ed25519.PrivateKeySize {
		t.Fatalf("private key: len=%d err=%v", len(priv), err)
	}
	pub, err := hex.DecodeString(pubHex)
	if err != nil || len(pub) != ed25519.PublicKeySize {
		t.Fatalf("public key: len=%d err=%v", len(pub), err)
	}
	if !ed25519.PublicKey(pub).Equal(ed25519.PrivateKey(priv).Public()) {
		t.Fatal("keypair halves do not match")
	}
}
