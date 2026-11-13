package detect

import (
	"bufio"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
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
	r := Result{Home: pathsHome()}
	r.SSHKeys = detectSSHKeys(r.Home)
	r.GitConfig = readGitConfig(r.Home)
	r.GitIdentity = parseGitIdentity(r.GitConfig)
	r.GHCli = detectGH()
	r.AZCli = detectAZ()
	r.GHLogin, r.GHEmail = detectGHUser()
	r.AZLogin, r.AZEmail = detectAZUser()
	return r
}

// pathsHome avoids import cycle
func pathsHome() string {
	h, _ := os.UserHomeDir()
	return h
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
	for _, line := range strings.Split(cfg, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "name") {
			// crude: name = value
		}
		if strings.HasPrefix(line, "name =") || strings.HasPrefix(line, "name=") {
			id.Name = valueAfterEquals(line)
		}
		if strings.HasPrefix(line, "email =") || strings.HasPrefix(line, "email=") {
			id.Email = valueAfterEquals(line)
		}
	}
	return id
}

func valueAfterEquals(s string) string {
	idx := strings.Index(s, "=")
	if idx < 0 {
		return ""
	}
	return strings.TrimSpace(s[idx+1:])
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
	_ = time.Second
	return info
}