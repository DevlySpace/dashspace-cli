package commands

import (
	"github.com/devlyspace/dashspace-cli/internal/commands/build"
	"github.com/spf13/cobra"
)

func NewBuildCmd() *cobra.Command {
	return build.NewBuildCmd()
}
