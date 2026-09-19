package licx

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
)

// GenerateKey creates a fresh Ed25519 keypair for signing license keys.
//
// privateB64 is the base64 64-byte private key (seed + public key), the
// format the issuing side stores. pubHex is the hex-encoded 32-byte public
// key, the format injected into verifiers via ldflags:
//
//	-X github.com/mizuchilabs/kata/licx.pubkey=<hex>
func GenerateKey() (privateB64, pubHex string, err error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return "", "", err
	}
	return base64.StdEncoding.EncodeToString(priv),
		hex.EncodeToString(pub), nil
}
