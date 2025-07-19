package main

import (
	"fmt"
	"os"

	"github.com/devlyspace/dashspace-cli/internal/commands"
	"github.com/devlyspace/dashspace-cli/internal/config"
	"github.com/spf13/cobra"
)

var version = "1.0.0"

func main() {
	rootCmd := &cobra.Command{
		Use:   "dashspace",
		Short: "CLI officiel DashSpace pour créer et publier des modules",
		Long: `CLI DashSpace permet aux développeurs de créer, tester et publier 
des modules pour l'écosystème DashSpace.`,
		Version: version,
	}

	config.InitConfig()

	rootCmd.AddCommand(commands.NewLoginCmd())
	rootCmd.AddCommand(commands.NewLogoutCmd())
	rootCmd.AddCommand(commands.NewWhoamiCmd())
	rootCmd.AddCommand(commands.NewCreateCmd())
	rootCmd.AddCommand(commands.NewPreviewCmd())
	rootCmd.AddCommand(commands.NewPublishCmd())
	rootCmd.AddCommand(commands.NewSearchCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
