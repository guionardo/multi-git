package paths

import (
	"os"
	"path/filepath"
	"testing"
)

func TestJoin(t *testing.T) {
	got := Join("a", "b", "c")
	if got != filepath.Join("a", "b", "c") {
		t.Fatalf("got %s", got)
	}
}

func TestGitPath(t *testing.T) {
	got := GitPath(filepath.Join("a", "b"))
	if got != "a/b" {
		t.Fatalf("got %s", got)
	}
}

func TestExpandHome(t *testing.T) {
	h := Home()
	got := ExpandHome("~/foo")
	if got != filepath.Join(h, "foo") {
		t.Fatalf("got %s", got)
	}
	if got := ExpandHome("/abs/path"); got != "/abs/path" {
		t.Fatalf("got %s", got)
	}
}

func TestHome(t *testing.T) {
	h := Home()
	if h == "" {
		t.Fatal("empty home")
	}
	if _, err := os.Stat(h); err != nil {
		t.Fatalf("home not statable: %v", err)
	}
}