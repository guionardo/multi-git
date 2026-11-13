package cmd

import (
	"bufio"
	"fmt"

	"github.com/guionardo/multi-git/internal/detect"
)

func runDetect(args []string) error {
	parseFlags(args)
	d := detect.Run()
	fmt.Println("Home:", d.Home)
	fmt.Println()
	fmt.Println("SSH keys:")
	if len(d.SSHKeys) == 0 {
		fmt.Println("  (none found in ~/.ssh)")
	}
	for _, k := range d.SSHKeys {
		priv := ""
		if k.HasPrivate {
			priv = " (private key present)"
		}
		fmt.Printf("  - %s [%s] %s%s\n", k.Name, k.Type, k.Comment, priv)
	}
	fmt.Println()
	fmt.Println("Existing Git identity:", d.GitIdentity.Name, d.GitIdentity.Email)
	if d.GHCli != nil {
		fmt.Println("gh CLI:", d.GHCli.Path)
		fmt.Println(d.GHCli.Detail)
		if d.GHLogin != "" {
			fmt.Println("  login:", d.GHLogin)
		}
		if d.GHEmail != "" {
			fmt.Println("  email:", d.GHEmail)
		}
	} else {
		fmt.Println("gh CLI: not found")
	}
	if d.AZCli != nil {
		fmt.Println("az CLI:", d.AZCli.Path)
		fmt.Println(d.AZCli.Detail)
		if d.AZLogin != "" {
			fmt.Println("  login:", d.AZLogin)
		}
		if d.AZEmail != "" {
			fmt.Println("  email:", d.AZEmail)
		}
	} else {
		fmt.Println("az CLI: not found")
	}
	_ = bufio.NewReader
	return nil
}