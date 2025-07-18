package commands

import (
	"fmt"
	"github.com/devlyspace/devly-cli/internal/api"
	"github.com/spf13/cobra"
)

func NewSearchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "search [query]",
		Short: "Rechercher des modules dans le store DashSpace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := args[0]
			return searchModules(query)
		},
	}
}

func searchModules(query string) error {
	fmt.Printf("🔍 Recherche de modules: %s\n\n", query)

	client := api.NewClient()
	modules, err := client.SearchModules(query)
	if err != nil {
		return fmt.Errorf("erreur recherche: %v", err)
	}

	if len(modules) == 0 {
		fmt.Println("❌ Aucun module trouvé")
		return nil
	}

	fmt.Printf("📦 %d modules trouvés:\n\n", len(modules))

	for _, module := range modules {
		fmt.Printf("• %s v%s\n", module.Name, module.Version)
		fmt.Printf("  %s\n", module.Description)
		fmt.Printf("  👤 %s\n", module.Author)
		fmt.Println()
	}

	return nil
}
