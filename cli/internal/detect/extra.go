package detect

import (
	"encoding/json"
	"os/exec"
	"strings"
)

func detectGHUser() (string, string) {
	out, err := exec.Command("gh", "api", "user").Output()
	if err != nil {
		return "", ""
	}
	var m map[string]any
	if json.Unmarshal(out, &m) != nil {
		return "", ""
	}
	login, _ := m["login"].(string)
	email, _ := m["email"].(string)
	return login, email
}

func detectAZUser() (string, string) {
	out, err := exec.Command("az", "account", "show", "-o", "json").Output()
	if err != nil {
		return "", ""
	}
	var m map[string]any
	if json.Unmarshal(out, &m) != nil {
		return "", ""
	}
	user, ok := m["user"].(map[string]any)
	if !ok || user == nil {
		return "", ""
	}
	name, _ := user["name"].(string)
	login := name
	email := ""
	if strings.Contains(name, "@") {
		email = name
	}
	return login, email
}