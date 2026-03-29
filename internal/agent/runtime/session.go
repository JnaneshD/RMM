package runtime

import (
	"bytes"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const (
	PinnedFingerprint = "C4:20:BE:D3:ED:17:AF:42:43:F8:45:10:F6:58:48:2F:11:B3:68:8E:C7:B6:E7:EB:2C:86:65:F4:02:8E:6A:A6"

	ServerHTTPS = "https://localhost:8080"
	ServerWSS   = "wss://localhost:8080"
)

func VerifyFingerprint(cs tls.ConnectionState) error {
	if len(cs.PeerCertificates) == 0 {
		return fmt.Errorf("no server certificate received")
	}

	// Compute SHA-256 of the raw DER cert bytes
	raw := cs.PeerCertificates[0].Raw
	sum := sha256.Sum256(raw)
	got := hex.EncodeToString(sum[:])

	// Normalise pinned value — strip colons, lowercase
	want := strings.ToLower(strings.ReplaceAll(PinnedFingerprint, ":", ""))

	if got != want {
		return fmt.Errorf("cert fingerprint mismatch\n  got:  %s\n  want: %s", got, want)
	}

	log.Println("[tls] server cert fingerprint OK")
	return nil
}

// pinnedTLSConfig returns a tls.Config that skips CA validation
// and does its own fingerprint check instead.
func PinnedTLSConfig() *tls.Config {
	return &tls.Config{
		// We skip the OS trust store — our fingerprint check is stricter.
		InsecureSkipVerify: true,
		VerifyConnection:   VerifyFingerprint,
	}
}

// buildHTTPClient returns an *http.Client with cert pinning.
// Used only for the POST /register call.
func BuildHTTPClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: PinnedTLSConfig(),
		},
		Timeout: 10 * time.Second,
	}
}

// buildWSDialer returns a *websocket.Dialer with cert pinning.
// Used for the persistent wss:// connection.
func BuildWSDialer() *websocket.Dialer {
	return &websocket.Dialer{
		TLSClientConfig:  PinnedTLSConfig(),
		HandshakeTimeout: 10 * time.Second,
	}
}

// -------------------------------------------------------------------
// Registration — HTTP POST /register
// Returns the session token to use for the WS connection.
// -------------------------------------------------------------------

func Register(client *http.Client, agentUUID string) (string, error) {
	payload, err := json.Marshal(map[string]string{
		"uuid":        agentUUID,
		"fingerprint": "localprint", // replace with real hw fingerprint later
	})
	if err != nil {
		return "", fmt.Errorf("marshal payload: %w", err)
	}

	resp, err := client.Post(ServerHTTPS+"/register", "application/json", bytes.NewReader(payload))
	if err != nil {
		// If fingerprint mismatch, the error surfaces here
		return "", fmt.Errorf("registration request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("register returned status %d", resp.StatusCode)
	}

	// NOTE: json.Decoder only populates exported (capitalised) fields.
	// Your original code had `session_token` (unexported) — fixed here.
	var result struct {
		SessionToken string `json:"session_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode register response: %w", err)
	}
	if result.SessionToken == "" {
		return "", fmt.Errorf("server returned empty session token")
	}

	return result.SessionToken, nil
}

// -------------------------------------------------------------------
// WebSocket connection — wss:// with cert pinning + token auth
// -------------------------------------------------------------------

func ConnectWS(dialer *websocket.Dialer, token string) (*websocket.Conn, error) {
	url := fmt.Sprintf("%s/ws/agent1?token=%s", ServerWSS, token)
	log.Printf("[ws] connecting to %s", url)

	conn, _, err := dialer.Dial(url, nil)
	if err != nil {
		// Fingerprint mismatch or network error surfaces here
		return nil, fmt.Errorf("ws dial failed: %w", err)
	}

	log.Println("[ws] connected — tunnel established and identity locked")
	return conn, nil
}
