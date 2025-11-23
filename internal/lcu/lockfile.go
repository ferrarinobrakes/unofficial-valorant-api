package lcu

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type LockfileData struct {
	Name     string
	PID      string
	Port     string
	Password string
	Protocol string
}

func ReadLockfile() (*LockfileData, error) {
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		return nil, fmt.Errorf("LOCALAPPDATA environment variable not set")
	}

	lockfilePath := filepath.Join(localAppData, "Riot Games", "Riot Client", "Config", "lockfile")

	file, err := os.Open(lockfilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open lockfile: %w (is Riot Client running?)", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		return nil, fmt.Errorf("lockfile is empty")
	}

	line := scanner.Text()
	parts := strings.Split(line, ":")
	if len(parts) != 5 {
		return nil, fmt.Errorf("invalid lockfile format: expected 5 parts, got %d", len(parts))
	}

	return &LockfileData{
		Name:     parts[0],
		PID:      parts[1],
		Port:     parts[2],
		Password: parts[3],
		Protocol: parts[4],
	}, nil
}

func (l *LockfileData) GetBaseURL() string {
	return fmt.Sprintf("https://127.0.0.1:%s", l.Port)
}
func (l *LockfileData) GetAuthHeader() string {
	return "Basic " + encodeBasicAuth("riot", l.Password)
}

// manual so no imports
// stolen from https://cs.opensource.google/go/go/+/master:src/encoding/base64/base64.go
func encodeBasicAuth(username, password string) string {
	auth := username + ":" + password
	const base64Table = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"

	result := make([]byte, 0, (len(auth)+2)/3*4)
	for i := 0; i < len(auth); i += 3 {
		b0 := auth[i]
		b1 := byte(0)
		b2 := byte(0)

		if i+1 < len(auth) {
			b1 = auth[i+1]
		}
		if i+2 < len(auth) {
			b2 = auth[i+2]
		}

		result = append(result, base64Table[b0>>2])
		result = append(result, base64Table[((b0&0x03)<<4)|(b1>>4)])

		if i+1 < len(auth) {
			result = append(result, base64Table[((b1&0x0f)<<2)|(b2>>6)])
		} else {
			result = append(result, '=')
		}

		if i+2 < len(auth) {
			result = append(result, base64Table[b2&0x3f])
		} else {
			result = append(result, '=')
		}
	}

	return string(result)
}
