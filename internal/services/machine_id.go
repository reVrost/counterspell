package services

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"sort"
	"strings"
)

var errNoMachineID = errors.New("no machine identifier found")

func stableMachineID() (string, error) {
	rawID, err := platformMachineID()
	if err != nil || rawID == "" {
		rawID = macAddressFallback()
	}
	if rawID == "" {
		host, _ := os.Hostname()
		rawID = fmt.Sprintf("host:%s|os:%s|arch:%s", host, runtime.GOOS, runtime.GOARCH)
	}
	return hashMachineID(rawID), nil
}

func hashMachineID(raw string) string {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	sum := sha256.Sum256([]byte(normalized))
	return "machine_" + hex.EncodeToString(sum[:])
}

func platformMachineID() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		return darwinMachineID()
	case "linux":
		return linuxMachineID()
	case "windows":
		return windowsMachineID()
	default:
		return "", errNoMachineID
	}
}

func linuxMachineID() (string, error) {
	return firstNonEmptyFile("/etc/machine-id", "/var/lib/dbus/machine-id")
}

func darwinMachineID() (string, error) {
	out, err := exec.Command("ioreg", "-rd1", "-c", "IOPlatformExpertDevice").Output()
	if err != nil {
		return "", err
	}
	// Example line: "IOPlatformUUID" = "XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX"
	re := regexp.MustCompile(`"IOPlatformUUID"\s*=\s*"([^"]+)"`)
	matches := re.FindSubmatch(out)
	if len(matches) < 2 {
		return "", errNoMachineID
	}
	return string(matches[1]), nil
}

func windowsMachineID() (string, error) {
	out, err := exec.Command("reg", "query", `HKLM\SOFTWARE\Microsoft\Cryptography`, "/v", "MachineGuid").Output()
	if err != nil {
		return "", err
	}
	// Example line: MachineGuid    REG_SZ    XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX
	fields := strings.Fields(string(out))
	if len(fields) < 3 {
		return "", errNoMachineID
	}
	return fields[len(fields)-1], nil
}

func firstNonEmptyFile(paths ...string) (string, error) {
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		value := strings.TrimSpace(string(b))
		if value != "" {
			return value, nil
		}
	}
	return "", errNoMachineID
}

func macAddressFallback() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return ""
	}
	macs := make([]string, 0, len(ifaces))
	for _, iface := range ifaces {
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		if len(iface.HardwareAddr) == 0 {
			continue
		}
		macs = append(macs, iface.HardwareAddr.String())
	}
	if len(macs) == 0 {
		return ""
	}
	sort.Strings(macs)
	return strings.Join(macs, ",")
}
