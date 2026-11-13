package config

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/guionardo/multi-git/internal/backup"
)

//go:embed templates/*
var templateFS embed.FS

type Data struct {
	PersonalName  string
	PersonalEmail string
	PersonalKey   string
	CorporateName  string
	CorporateEmail string
	CorporateKey   string
}

func render(name string, data Data) (string, error) {
	b, err := templateFS.ReadFile("templates/" + name)
	if err != nil {
		return "", err
	}
	t, err := template.New(name).Parse(string(b))
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

type FileSpec struct {
	Target string
	Template string
	Mode   os.FileMode
}

func Specs(home string) []FileSpec {
	return []FileSpec{
		{Target: filepath.Join(home, ".gitconfig"), Template: "gitconfig.tmpl", Mode: 0644},
		{Target: filepath.Join(home, ".gitconfig-personal"), Template: "gitconfig.personal.tmpl", Mode: 0644},
		{Target: filepath.Join(home, ".gitconfig-work"), Template: "gitconfig.work.tmpl", Mode: 0644},
		{Target: filepath.Join(home, ".ssh", "config"), Template: "ssh_config.tmpl", Mode: 0600},
	}
}

// Apply renders and writes all specs. Returns list of written paths.
func Apply(data Data, home string, dryRun, yes bool, confirm func(string) bool) ([]string, error) {
	var written []string
	for _, spec := range Specs(home) {
		content, err := render(spec.Template, data)
		if err != nil {
			return written, err
		}
		if backup.Exists(spec.Target) {
			if !yes && !confirm("Overwrite "+spec.Target+"?") {
				fmt.Println("skipped:", spec.Target)
				continue
			}
			bp, err := backup.Backup(spec.Target)
			if err != nil {
				return written, fmt.Errorf("backup %s: %w", spec.Target, err)
			}
			if bp != "" {
				fmt.Println("backup:", bp)
			}
		} else {
			if err := os.MkdirAll(filepath.Dir(spec.Target), 0755); err != nil {
				return written, fmt.Errorf("create dir for %s: %w", spec.Target, err)
			}
		}
		if dryRun {
			fmt.Println("dry-run write:", spec.Target)
			written = append(written, spec.Target)
			continue
		}
		if err := os.WriteFile(spec.Target, []byte(content), spec.Mode); err != nil {
			return written, err
		}
		fmt.Println("written:", spec.Target)
		written = append(written, spec.Target)
	}
	// allowed_signers
	as := filepath.Join(home, ".ssh", "allowed_signers")
	personalPub := readPubKey(filepath.Join(home, ".ssh", data.PersonalKey+".pub"))
	corporatePub := readPubKey(filepath.Join(home, ".ssh", data.CorporateKey+".pub"))
	if personalPub == "" {
		return written, fmt.Errorf("personal public key not found: %s", data.PersonalKey+".pub")
	}
	if corporatePub == "" {
		return written, fmt.Errorf("corporate public key not found: %s", data.CorporateKey+".pub")
	}
	linePersonal := fmt.Sprintf("%s %s\n", data.PersonalEmail, personalPub)
	lineCorp := fmt.Sprintf("%s %s\n", data.CorporateEmail, corporatePub)
	content := linePersonal + lineCorp
	if backup.Exists(as) {
		if !yes && !confirm("Overwrite "+as+"?") {
			fmt.Println("skipped:", as)
		} else {
			bp, err := backup.Backup(as)
			if err != nil {
				return written, fmt.Errorf("backup %s: %w", as, err)
			}
			if bp != "" {
				fmt.Println("backup:", bp)
			}
			if !dryRun {
				if err := os.WriteFile(as, []byte(content), 0600); err != nil {
					return written, fmt.Errorf("write %s: %w", as, err)
				}
				fmt.Println("written:", as)
				written = append(written, as)
			} else {
				fmt.Println("dry-run write:", as)
				written = append(written, as)
			}
		}
	} else {
		if !dryRun {
			if err := os.MkdirAll(filepath.Dir(as), 0755); err != nil {
				return written, fmt.Errorf("create dir for %s: %w", as, err)
			}
			if err := os.WriteFile(as, []byte(content), 0600); err != nil {
				return written, fmt.Errorf("write %s: %w", as, err)
			}
			fmt.Println("written:", as)
			written = append(written, as)
		} else {
			fmt.Println("dry-run write:", as)
			written = append(written, as)
		}
	}
	return written, nil
}

func readPubKey(p string) string {
	b, err := os.ReadFile(p)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}