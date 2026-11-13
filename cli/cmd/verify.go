package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/guionardo/multi-git/internal/paths"
)

func runVerify(args []string) error {
	parseFlags(args)
	home := paths.Home()
	required := []string{
		filepath.Join(home, ".gitconfig"),
		filepath.Join(home, ".gitconfig-personal"),
		filepath.Join(home, ".gitconfig-work"),
		filepath.Join(home, ".ssh", "config"),
		filepath.Join(home, ".ssh", "allowed_signers"),
	}
	ok := true
	for _, f := range required {
		if _, err := os.Stat(f); err != nil {
			fmt.Println("MISSING:", f)
			ok = false
		} else {
			// placeholder check
			b, _ := os.ReadFile(f)
			if strings.Contains(string(b), "{{") {
				fmt.Println("PLACEHOLDER REMAINS:", f)
				ok = false
			} else {
				fmt.Println("OK:", f)
			}
		}
	}

	// Test identity resolution in temp repos
	for _, tc := range []struct {
		dir  string
		want string
	}{
		{filepath.Join(home, "personal", ".multi-git-verify"), "personal"},
		{filepath.Join(home, "work", ".multi-git-verify"), "corporate"},
	} {
		if err := os.MkdirAll(tc.dir, 0755); err != nil {
			fmt.Printf("  %s: could not create dir: %v\n", tc.dir, err)
			ok = false
			continue
		}
		initCmd := exec.Command("git", "init", "-q")
		initCmd.Dir = tc.dir
		if err := initCmd.Run(); err != nil {
			fmt.Printf("  %s: git init failed: %v\n", tc.dir, err)
			ok = false
			_ = os.RemoveAll(tc.dir)
			continue
		}
		// query email
		out, err := exec.Command("git", "-C", tc.dir, "config", "user.email").Output()
		email := strings.TrimSpace(string(out))
		if err != nil || email == "" {
			fmt.Printf("  %s: could not resolve identity\n", tc.dir)
			ok = false
		} else {
			fmt.Printf("  %s -> %s\n", tc.dir, email)
		}
		_ = os.RemoveAll(tc.dir)
	}

	if !ok {
		return fmt.Errorf("verification failed")
	}
	fmt.Println("Verification passed")
	return nil
}