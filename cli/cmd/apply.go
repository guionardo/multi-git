package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/guionardo/multi-git/internal/config"
	"github.com/guionardo/multi-git/internal/detect"

	"github.com/spf13/cobra"
)

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Interactively generate/update Git and SSH configuration",
	RunE:  func(cmd *cobra.Command, args []string) error { return runApply() },
}

func runApplyArgs(args []string) error {
	return runApply()
}

func runApply() error {
	reader := bufio.NewReader(os.Stdin)
	d := detect.Run()

	fmt.Println("=== multi-git apply ===")
	fmt.Println("Detected home:", d.Home)

	// Defaults
	personalKey := "id_ed25519"
	corporateKey := "id_rsa"
	if !keyExists(d.SSHKeys, personalKey) {
		if len(d.SSHKeys) > 0 {
			personalKey = d.SSHKeys[0].Name
		}
	}
	if !keyExists(d.SSHKeys, corporateKey) {
		if len(d.SSHKeys) > 1 {
			corporateKey = d.SSHKeys[1].Name
		} else if len(d.SSHKeys) > 0 {
			corporateKey = d.SSHKeys[0].Name
		}
	}

	personalName := d.GitIdentity.Name
	if personalName == "" && d.GHLogin != "" {
		personalName = d.GHLogin
	}
	if personalName == "" {
		personalName = "Your Name"
	}
	personalEmail := d.GitIdentity.Email
	if personalEmail == "" && d.GHEmail != "" {
		personalEmail = d.GHEmail
	}
	if personalEmail == "" {
		personalEmail = "you@example.com"
	}
	corporateName := personalName
	if d.AZLogin != "" && !strings.Contains(d.AZLogin, "@") {
		corporateName = d.AZLogin
	}
	corporateEmail := "you@company.com"
	if d.AZEmail != "" {
		corporateEmail = d.AZEmail
	}

	// Try gh/az hints (best effort)
	if d.GHCli != nil && d.GHCli.Detail != "" {
		// keep defaults; user confirms
	}
	if d.AZCli != nil && d.AZCli.Detail != "" {
		// keep defaults
	}

	fmt.Println("\nPersonal identity:")
	data := config.Data{}
	data.PersonalName = prompt(reader, "Personal name", personalName)
	data.PersonalEmail = prompt(reader, "Personal email", personalEmail)
	data.PersonalKey = prompt(reader, "Personal SSH key name", personalKey)

	fmt.Println("\nCorporate identity:")
	data.CorporateName = prompt(reader, "Corporate name", corporateName)
	data.CorporateEmail = prompt(reader, "Corporate email", corporateEmail)
	data.CorporateKey = prompt(reader, "Corporate SSH key name", corporateKey)

	fmt.Println("\nSummary:")
	fmt.Println("  Personal :", data.PersonalName, "<"+data.PersonalEmail+">", data.PersonalKey)
	fmt.Println("  Corporate:", data.CorporateName, "<"+data.CorporateEmail+">", data.CorporateKey)

	if !confirm(reader, "Apply this configuration?") {
		fmt.Println("aborted")
		return nil
	}

	home := d.Home
	// Ensure dirs
	for _, dir := range []string{filepath.Join(home, "personal"), filepath.Join(home, "work")} {
		if _, err := os.Stat(dir); err != nil {
			if confirm(reader, "Create directory "+dir+"?") {
				if !Flags.DryRun {
					if err := os.MkdirAll(dir, 0755); err != nil {
						return fmt.Errorf("create dir %s: %w", dir, err)
					}
				}
				fmt.Println("ensure dir:", dir)
			}
		}
	}

	written, err := config.Apply(data, home, Flags.DryRun, Flags.Yes, func(label string) bool {
		return confirm(reader, label)
	})
	if err != nil {
		return err
	}
	fmt.Println("\nWritten files:", len(written))
	fmt.Println("Next: move personal repos under ~/personal and corporate repos under ~/work.")
	fmt.Println("Run 'multi-git verify' to validate.")
	return nil
}

func keyExists(keys []detect.SSHKey, name string) bool {
	for _, k := range keys {
		if k.Name == name {
			return true
		}
	}
	return false
}

