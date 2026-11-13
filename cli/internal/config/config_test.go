package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestApplyDryRun(t *testing.T) {
	home := t.TempDir()
	sshDir := filepath.Join(home, ".ssh")
	if err := os.MkdirAll(sshDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sshDir, "id_ed25519.pub"), []byte("ssh-ed25519 AAAA test"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sshDir, "id_rsa.pub"), []byte("ssh-rsa AAAA test"), 0644); err != nil {
		t.Fatal(err)
	}
	data := Data{
		PersonalName:   "Alice",
		PersonalEmail:  "alice@example.com",
		PersonalKey:    "id_ed25519",
		CorporateName:  "Alice Corp",
		CorporateEmail: "alice@corp.com",
		CorporateKey:   "id_rsa",
	}
	written, err := Apply(data, home, true, true, func(string) bool { return true })
	if err != nil {
		t.Fatalf("Apply dry-run: %v", err)
	}
	if len(written) == 0 {
		t.Fatal("expected written files")
	}
	// dry-run should not create .gitconfig
	if _, err := os.Stat(filepath.Join(home, ".gitconfig")); err == nil {
		t.Fatal("dry-run should not write .gitconfig")
	}
}

func TestApplyWritesFiles(t *testing.T) {
	home := t.TempDir()
	sshDir := filepath.Join(home, ".ssh")
	if err := os.MkdirAll(sshDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sshDir, "id_ed25519.pub"), []byte("ssh-ed25519 AAAA test"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sshDir, "id_rsa.pub"), []byte("ssh-rsa AAAA test"), 0644); err != nil {
		t.Fatal(err)
	}
	data := Data{
		PersonalName:   "Alice",
		PersonalEmail:  "alice@example.com",
		PersonalKey:    "id_ed25519",
		CorporateName:  "Alice Corp",
		CorporateEmail: "alice@corp.com",
		CorporateKey:   "id_rsa",
	}
	written, err := Apply(data, home, false, true, func(string) bool { return true })
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if len(written) < 5 {
		t.Fatalf("expected >=5 written, got %d", len(written))
	}
	// verify no placeholders
	for _, f := range []string{".gitconfig", ".gitconfig-personal", ".gitconfig-work"} {
		b, _ := os.ReadFile(filepath.Join(home, f))
		if strings.Contains(string(b), "{{") {
			t.Fatalf("%s has placeholder", f)
		}
	}
}

func TestRenderNoPlaceholders(t *testing.T) {
	data := Data{
		PersonalName:  "Alice",
		PersonalEmail: "alice@example.com",
		PersonalKey:   "id_ed25519",
		CorporateName:  "Alice Corp",
		CorporateEmail: "alice@corp.com",
		CorporateKey:   "id_rsa",
	}
	for _, name := range []string{"gitconfig.tmpl", "gitconfig.personal.tmpl", "gitconfig.work.tmpl", "ssh_config.tmpl"} {
		out, err := render(name, data)
		if err != nil {
			t.Fatalf("render %s: %v", name, err)
		}
		if strings.Contains(out, "{{") {
			t.Fatalf("render %s still has placeholder: %s", name, out)
		}
	}
}