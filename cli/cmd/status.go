package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/guionardo/multi-git/internal/detect"
	"github.com/guionardo/multi-git/internal/paths"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current multi-git setup status",
	RunE:  func(cmd *cobra.Command, args []string) error { return runStatus() },
}

func runStatus() error {
	home := paths.Home()
	fmt.Println("Home:", home)
	files := []string{".gitconfig", ".gitconfig-personal", ".gitconfig-work", ".ssh/config", ".ssh/allowed_signers"}
	for _, f := range files {
		p := filepath.Join(home, f)
		if _, err := os.Stat(p); err != nil {
			fmt.Println("MISSING:", p)
		} else {
			fmt.Println("OK     :", p)
		}
	}
	d := detect.Run()
	fmt.Println()
	fmt.Println("SSH keys:", len(d.SSHKeys))
	for _, k := range d.SSHKeys {
		fmt.Printf("  - %s [%s]\n", k.Name, k.Type)
	}
	fmt.Println("Git identity:", d.GitIdentity.Name, d.GitIdentity.Email)
	return nil
}