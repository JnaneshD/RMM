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

const uuidFile = "agent.uuid"

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
