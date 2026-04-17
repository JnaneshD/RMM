package runtime

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"runtime"
	"strings"

	"github.com/google/uuid"
)

func HardwareFingerprint() (string, error) {
	mac, err := PrimaryMac()
	if err != nil {
		return "", fmt.Errorf("get MAC: %w", err)
	}

	hostname, err := os.Hostname()
	if err != nil {
		return "", fmt.Errorf("get hostname: %w", err)
	}

	// Now lets add these into one single string
	raw := strings.Join([]string{
		mac,
		strings.ToLower(hostname),
		runtime.GOOS,
		runtime.GOARCH,
	}, "|")

	sum := sha256.Sum256([]byte(raw))

	return hex.EncodeToString(sum[:]), nil
}

func PrimaryMac() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range ifaces {
		// Skip loopback, down interfaces, and those without a MAC
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		if len(iface.HardwareAddr) == 0 {
			continue
		}
		return iface.HardwareAddr.String(), nil
	}

	return "", fmt.Errorf("no suitable network interface found")
}

const (
	uuidFile  = "agent.uuid"
	tokenFile = "agent.token"
	urlFile   = "agent.url"
)

// AgentUUID returns the persistent agent UUID.
// First run: generates a new UUID and saves it to disk.
// Subsequent runs: reads the saved UUID from disk.
// This ensures the agent has the same identity across restarts.
func AgentUUID() (string, error) {
	// Try to read existing UUID from disk
	data, err := os.ReadFile(uuidFile)
	if err == nil {
		id := strings.TrimSpace(string(data))
		if id != "" {
			return id, nil
		}
	}

	// Not found — generate a new one and persist it
	id := uuid.New().String()
	if err := os.WriteFile(uuidFile, []byte(id), 0600); err != nil {
		return "", fmt.Errorf("save agent uuid: %w", err)
	}

	return id, nil
}

// LoadAgentToken reads the saved token from disk.
func LoadAgentToken() (string, error) {
	data, err := os.ReadFile(tokenFile)
	if err == nil {
		token := strings.TrimSpace(string(data))
		if token != "" {
			return token, nil
		}
	}
	return "", fmt.Errorf("token not found")
}

// SaveAgentToken persists the session token to disk.
func SaveAgentToken(token string) error {
	if err := os.WriteFile(tokenFile, []byte(token), 0600); err != nil {
		return fmt.Errorf("save agent token: %w", err)
	}
	return nil
}

// ClearAgentToken removes the token from disk.
func ClearAgentToken() error {
	return os.Remove(tokenFile)
}

// LoadServerURL reads the saved server URL from disk.
func LoadServerURL() (string, error) {
	data, err := os.ReadFile(urlFile)
	if err == nil {
		url := strings.TrimSpace(string(data))
		if url != "" {
			return url, nil
		}
	}
	return "", fmt.Errorf("server url not found")
}

// SaveServerURL persists the server URL to disk.
func SaveServerURL(url string) error {
	if err := os.WriteFile(urlFile, []byte(url), 0600); err != nil {
		return fmt.Errorf("save server url: %w", err)
	}
	return nil
}
