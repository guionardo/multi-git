package detect

import "testing"

func TestParseGitIdentity(t *testing.T) {
	cfg := `[user]
    name = Alice
    email = alice@example.com
[core]
    autocrlf = input
[user "x"]
    name = ignored
`
	id := parseGitIdentity(cfg)
	if id.Name != "Alice" {
		t.Fatalf("name = %q, want Alice", id.Name)
	}
	if id.Email != "alice@example.com" {
		t.Fatalf("email = %q, want alice@example.com", id.Email)
	}
}