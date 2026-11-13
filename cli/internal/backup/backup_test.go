package backup

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExists(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "f")
	if Exists(p) {
		t.Fatal("should not exist")
	}
	if err := os.WriteFile(p, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	if !Exists(p) {
		t.Fatal("should exist")
	}
}

func TestBackupAndRestore(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "f")
	if err := os.WriteFile(p, []byte("original"), 0644); err != nil {
		t.Fatal(err)
	}
	bp, err := Backup(p)
	if err != nil {
		t.Fatal(err)
	}
	if bp == "" {
		t.Fatal("expected backup path")
	}
	if _, err := os.Stat(bp); err != nil {
		t.Fatal("backup file missing")
	}
	// overwrite
	if err := os.WriteFile(p, []byte("changed"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := Restore(p, bp); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(p)
	if string(b) != "original" {
		t.Fatalf("restored content = %s", b)
	}
}

func TestBackupMissing(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "nope")
	bp, err := Backup(p)
	if err != nil {
		t.Fatal(err)
	}
	if bp != "" {
		t.Fatalf("expected empty backup path, got %s", bp)
	}
}