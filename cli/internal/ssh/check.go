package ssh

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// KeyPairMatch reports whether the private key at privPath produces the same
// public key as the contents of pubPath.
func KeyPairMatch(privPath, pubPath string) bool {
	if _, err := os.Stat(privPath); err != nil {
		return false
	}
	pubBytes, err := os.ReadFile(pubPath)
	if err != nil {
		return false
	}
	out, err := exec.Command("ssh-keygen", "-y", "-f", privPath).Output()
	if err != nil {
		return false
	}
	got := strings.TrimSpace(string(out))
	want := strings.TrimSpace(string(pubBytes))
	// compare only type + base64 (ignore comment)
	gotFields := strings.Fields(got)
	wantFields := strings.Fields(want)
	if len(gotFields) < 2 || len(wantFields) < 2 {
		return false
	}
	return gotFields[0] == wantFields[0] && gotFields[1] == wantFields[1]
}

// HostCheck runs `ssh -T` against host using the given key and reports success.
func HostCheck(host, keyPath string, timeout time.Duration) (bool, string) {
	ctx := exec.Command("ssh",
		"-o", "BatchMode=yes",
		"-o", "StrictHostKeyChecking=accept-new",
		"-o", "ConnectTimeout=10",
		"-i", keyPath,
		"-T", "git@"+host,
	)
	// Use a temp approach with timeout via process kill.
	done := make(chan struct{})
	var stdout, stderr bytes.Buffer
	ctx.Stdout = &stdout
	ctx.Stderr = &stderr
	if err := ctx.Start(); err != nil {
		return false, err.Error()
	}
	go func() {
		_ = ctx.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(timeout):
		_ = ctx.Process.Kill()
		return false, "timeout"
	}
	combined := strings.TrimSpace(stdout.String() + stderr.String())
	// ssh -T exits 1 on success for many hosts (e.g. GitHub), so check output text.
	lower := strings.ToLower(combined)
	if strings.Contains(lower, "successfully authenticated") ||
		strings.Contains(lower, "welcome") ||
		strings.Contains(lower, "shell access") ||
		strings.Contains(lower, "you've successfully authenticated") {
		return true, combined
	}
	// If exit code 0, treat as success.
	if ctx.ProcessState != nil && ctx.ProcessState.Success() {
		return true, combined
	}
	return false, combined
}

func KeyPermissions(path string) (bool, string) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err.Error()
	}
	mode := info.Mode().Perm()
	if mode&0077 != 0 {
		return false, fmt.Sprintf("permissions too open: %o (want 600)", mode)
	}
	return true, "ok"
}

func ExpandPath(home, p string) string {
	if strings.HasPrefix(p, "~/") {
		return filepath.Join(home, p[2:])
	}
	return p
}