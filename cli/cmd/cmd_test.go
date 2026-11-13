package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func runCmd(t *testing.T, args ...string) (string, error) {
	t.Helper()
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	os.Stderr = w
	rootCmd.SetOut(w)
	rootCmd.SetErr(w)
	rootCmd.SetArgs(args)
	runErr := rootCmd.Execute()
	w.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr
	rootCmd.SetArgs(nil)
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	r.Close()
	return buf.String(), runErr
}

func TestHelpCommand(t *testing.T) {
	out, err := runCmd(t, "help")
	if err != nil {
		t.Fatalf("help: %v", err)
	}
	if !contains(out, "multi-git") {
		t.Fatalf("help output missing title: %s", out)
	}
}

func TestDetectCommand(t *testing.T) {
	out, err := runCmd(t, "detect")
	if err != nil {
		t.Fatalf("detect: %v", err)
	}
	if !contains(out, "Home:") {
		t.Fatalf("detect output missing Home: %s", out)
	}
}

func TestStatusCommand(t *testing.T) {
	out, err := runCmd(t, "status")
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if !contains(out, "Home:") {
		t.Fatalf("status output missing Home: %s", out)
	}
}

func TestVerifyCommandMissingFiles(t *testing.T) {
	tmp := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmp)
	defer os.Setenv("HOME", oldHome)
	out, err := runCmd(t, "verify")
	if err == nil {
		t.Log("verify unexpectedly succeeded")
	}
	if !contains(out, "MISSING") {
		t.Fatalf("expected MISSING in output: %s", out)
	}
	_ = filepath.Join
}

func TestUnknownCommand(t *testing.T) {
	_, err := runCmd(t, "nope")
	if err == nil {
		t.Fatal("expected error for unknown command")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (func() bool {
		for i := 0; i+len(sub) <= len(s); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	})()
}