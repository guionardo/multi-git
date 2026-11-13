package config

import (
	"strings"
	"testing"
)

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