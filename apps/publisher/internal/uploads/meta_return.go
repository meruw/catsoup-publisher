package uploads

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"net/url"
)

// Meta registers one public HTTPS URI. The browser uses its fragment to pair
// the tab with the local listener; the pairing key is never sent to a server.

type metaReturn struct {
	key  []byte
	aead cipher.AEAD
}

func newMetaReturn() (*metaReturn, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &metaReturn{key: key, aead: aead}, nil
}
func (m *metaReturn) launch(redirect, port, state string, cfg desktopOAuth) string {
	fragment := url.Values{"app": {cfg.id}, "config": {cfg.scope}, "port": {port}, "state": {state}, "key": {base64.RawURLEncoding.EncodeToString(m.key)}}
	fragment.Set("return_uri", redirect)
	return redirect + "#" + fragment.Encode()
}
func (m *metaReturn) code(blob, state string) (string, bool) {
	if len(blob) > 16384 {
		return "", false
	}
	raw, err := base64.RawURLEncoding.DecodeString(blob)
	if err != nil || len(raw) < m.aead.NonceSize()+m.aead.Overhead() {
		return "", false
	}
	n := m.aead.NonceSize()
	plain, err := m.aead.Open(nil, raw[:n], raw[n:], []byte(state))
	return string(plain), err == nil
}
