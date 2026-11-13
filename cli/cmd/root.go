package cmd

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

type GlobalFlags struct {
	DryRun bool
	Yes    bool
}

var Flags GlobalFlags

var rootCmd = &cobra.Command{
	Use:   "multi-git",
	Short: "Configure multiple Git identities on one machine",
	Long: `multi-git helps developers safely use different Git identities
(personal and work) on a single machine using conditional includes,
SSH keys, and signed commits.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runApply()
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&Flags.DryRun, "dry-run", false, "show changes without writing files")
	rootCmd.PersistentFlags().BoolVar(&Flags.Yes, "yes", false, "do not prompt for confirmation")

	rootCmd.AddCommand(detectCmd)
	rootCmd.AddCommand(applyCmd)
	rootCmd.AddCommand(verifyCmd)
	rootCmd.AddCommand(wizardCmd)
	rootCmd.AddCommand(statusCmd)
}

func printHelp() {
	fmt.Println(rootCmd.UsageString())
}

// ---- shared helpers ----

func prompt(reader *bufio.Reader, label, def string) string {
	if def != "" {
		fmt.Printf("%s [%s]: ", label, def)
	} else {
		fmt.Printf("%s: ", label)
	}
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(line)
	if line == "" {
		return def
	}
	return line
}

func confirm(reader *bufio.Reader, label string) bool {
	if Flags.Yes {
		return true
	}
	fmt.Printf("%s [y/N]: ", label)
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(strings.ToLower(line))
	return line == "y" || line == "yes"
}