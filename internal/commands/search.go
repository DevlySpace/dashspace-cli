package commands

import (
	"fmt"

	"github.com/devlyspace/devly-cli/internal/api"
	"github.com/spf13/cobra"
)

func NewSearchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "search [query]",
		Short: "Search modules in DashSpace store",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := args[0]
			return searchModules(query)
		},
	}
}

func searchModules(query string) error {
	fmt.Printf("🔍 Searching modules: %s\n\n", query)

	client := api.NewClient()
	modules, err := client.SearchModules(query)
	if err != nil {
		return fmt.Errorf("search error: %v", err)
	}

	if len(modules) == 0 {
		fmt.Println("❌ No modules found")
		return nil
	}

	fmt.Printf("📦 %d modules found:\n\n", len(modules))

	for _, module := range modules {
		fmt.Printf("• %s v%s\n", module.Name, module.Version)
		fmt.Printf("  %s\n", module.Description)
		fmt.Printf("  👤 %s\n", module.Author)
		if len(module.Tags) > 0 {
			fmt.Printf("  🏷️  Tags: %v\n", module.Tags)
		}
		fmt.Println()
	}

	return nil
}
