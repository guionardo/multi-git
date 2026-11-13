package cmd

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
)

type GlobalFlags struct {
	DryRun bool
	Yes    bool
}

var Flags GlobalFlags

func Execute() error {
	if len(os.Args) < 2 {
		return runApply()
	}
	switch os.Args[1] {
	case "detect":
		return runDetect(os.Args[2:])
	case "apply":
		return runApplyArgs(os.Args[2:])
	case "verify":
		return runVerify(os.Args[2:])
	case "help", "-h", "--help":
		printHelp()
		return nil
	default:
		return fmt.Errorf("unknown command %q (use: detect, apply, verify, help)", os.Args[1])
	}
}

func parseFlags(args []string) {
	fs := flag.NewFlagSet("multi-git", flag.ExitOnError)
	fs.BoolVar(&Flags.DryRun, "dry-run", false, "show changes without writing files")
	fs.BoolVar(&Flags.Yes, "yes", false, "answer yes to all confirmations")
	fs.Parse(args)
}

func printHelp() {
	fmt.Println(`multi-git — configure multiple Git identities on one machine

Usage:
  multi-git [command] [flags]

Commands:
  detect   Show detected SSH keys, CLIs and existing Git config
  apply    Interactively generate/update Git and SSH configuration
  verify   Validate generated configuration
  help     Show this help

Flags:
  --dry-run   Do not write files
  --yes       Do not prompt for confirmation`)
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