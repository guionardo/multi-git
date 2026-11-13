package detect

import (
	"bufio"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/guionardo/multi-git/internal/paths"
)

type SSHKey struct {
	Name       string
	Private    string
	Public     string
	Type       string
	Comment    string
	HasPrivate bool
}

type CLIInfo struct {
	Name    string
	Path    string
	Version string
	Detail  string
}

type GitIdentity struct {
	Name  string
	Email string
}

type Result struct {
	Home        string
	SSHKeys     []SSHKey
	GitConfig   string
	GitIdentity GitIdentity
	GHCli       *CLIInfo
	AZCli       *CLIInfo
	GHLogin     string
	GHEmail     string
	AZLogin     string
	AZEmail     string
}

func Run() Result {
	r := Result{Home: paths.Home()}
	r.SSHKeys = detectSSHKeys(r.Home)
	r.GitConfig = readGitConfig(r.Home)
	r.GitIdentity = parseGitIdentity(r.GitConfig)
	r.GHCli = detectGH()
	r.AZCli = detectAZ()
	r.GHLogin, r.GHEmail = detectGHUser()
	r.AZLogin, r.AZEmail = detectAZUser()
	return r
}

func detectSSHKeys(home string) []SSHKey {
	dir := filepath.Join(home, ".ssh")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var keys []SSHKey
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".pub") {
			continue
		}
		base := strings.TrimSuffix(name, ".pub")
		pubPath := filepath.Join(dir, name)
		privPath := filepath.Join(dir, base)
		f, err := os.Open(pubPath)
		if err != nil {
			continue
		}
		sc := bufio.NewScanner(f)
		var line string
		if sc.Scan() {
			line = strings.TrimSpace(sc.Text())
		}
		f.Close()
		fields := strings.Fields(line)
		k := SSHKey{Name: base, Private: privPath, Public: pubPath}
		if len(fields) >= 1 {
			k.Type = fields[0]
		}
		if len(fields) >= 3 {
			k.Comment = strings.Join(fields[2:], " ")
		}
		if _, err := os.Stat(privPath); err == nil {
			k.HasPrivate = true
		}
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i].Name < keys[j].Name })
	return keys
}

func readGitConfig(home string) string {
	b, err := os.ReadFile(filepath.Join(home, ".gitconfig"))
	if err != nil {
		return ""
	}
	return string(b)
}

func parseGitIdentity(cfg string) GitIdentity {
	id := GitIdentity{}
	section := ""
	for _, raw := range strings.Split(cfg, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.ToLower(strings.TrimSpace(line[1 : len(line)-1]))
			continue
		}
		if section != "user" {
			continue
		}
		key, val, ok := splitKeyValue(line)
		if !ok {
			continue
		}
		switch strings.ToLower(key) {
		case "name":
			id.Name = val
		case "email":
			id.Email = val
		}
	}
	return id
}

func splitKeyValue(line string) (string, string, bool) {
	idx := strings.Index(line, "=")
	if idx < 0 {
		return "", "", false
	}
	return strings.TrimSpace(line[:idx]), strings.TrimSpace(line[idx+1:]), true
}

func detectGH() *CLIInfo {
	path, err := exec.LookPath("gh")
	if err != nil {
		return nil
	}
	info := &CLIInfo{Name: "gh", Path: path}
	if out, err := exec.Command("gh", "--version").Output(); err == nil {
		info.Version = strings.TrimSpace(string(out))
	}
	// auth status
	if out, err := exec.Command("gh", "auth", "status").CombinedOutput(); err == nil {
		info.Detail = strings.TrimSpace(string(out))
	} else {
		info.Detail = strings.TrimSpace(string(out))
	}
	return info
}

func detectAZ() *CLIInfo {
	path, err := exec.LookPath("az")
	if err != nil {
		return nil
	}
	info := &CLIInfo{Name: "az", Path: path}
	if out, err := exec.Command("az", "--version").Output(); err == nil {
		lines := strings.SplitN(string(out), "\n", 2)
		if len(lines) > 0 {
			info.Version = strings.TrimSpace(lines[0])
		}
	}
	if out, err := exec.Command("az", "account", "show", "-o", "json").Output(); err == nil {
		var m map[string]any
		if json.Unmarshal(out, &m) == nil {
			if u, ok := m["user"].(map[string]any); ok {
				if n, ok := u["name"].(string); ok {
					info.Detail = "account: " + n
				}
			}
		}
	} else {
		info.Detail = "not logged in"
	}
	return info
}
