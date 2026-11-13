package ssh

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestKeyPermissions(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "key")
	if err := os.WriteFile(p, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	ok, _ := KeyPermissions(p)
	if !ok {
		t.Fatal("expected ok for 0600")
	}
	if err := os.Chmod(p, 0644); err != nil {
		t.Fatal(err)
	}
	ok, _ = KeyPermissions(p)
	if ok {
		t.Fatal("expected not ok for 0644")
	}
}

func TestExpandPath(t *testing.T) {
	if got := ExpandPath("/home/u", "~/x"); got != "/home/u/x" {
		t.Fatalf("got %s", got)
	}
	if got := ExpandPath("/home/u", "/abs"); got != "/abs" {
		t.Fatalf("got %s", got)
	}
}

func TestKeyPairMatch(t *testing.T) {
	if _, err := exec.LookPath("ssh-keygen"); err != nil {
		t.Skip("ssh-keygen not available")
	}
	dir := t.TempDir()
	priv := filepath.Join(dir, "id_test")
	pub := priv + ".pub"
	cmd := exec.Command("ssh-keygen", "-t", "ed25519", "-N", "", "-f", priv, "-C", "test")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("ssh-keygen: %v: %s", err, out)
	}
	if !KeyPairMatch(priv, pub) {
		t.Fatal("expected key pair to match")
	}
	// mismatched pub
	other := filepath.Join(dir, "other.pub")
	if err := os.WriteFile(other, []byte("ssh-ed25519 AAAA nope"), 0644); err != nil {
		t.Fatal(err)
	}
	if KeyPairMatch(priv, other) {
		t.Fatal("expected mismatch")
	}
}