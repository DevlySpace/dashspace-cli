package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/devlyspace/devly-cli/internal/templates"
	"github.com/spf13/cobra"
)

func NewCreateCmd() *cobra.Command {
	var templateType string
	var providers []string

	cmd := &cobra.Command{
		Use:   "create [nom-du-module]",
		Short: "Créer un nouveau module DashSpace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			moduleName := args[0]
			return createModule(moduleName, templateType, providers)
		},
	}

	cmd.Flags().StringVarP(&templateType, "template", "t", "react", "Type de template (react, vanilla, chart)")
	cmd.Flags().StringSliceVarP(&providers, "providers", "p", []string{}, "Providers requis (github,slack,asana)")

	return cmd
}

func createModule(name, templateType string, providers []string) error {
	fmt.Printf("🚀 Création du module '%s'\n", name)

	// Vérifier si le dossier existe déjà
	if _, err := os.Stat(name); err == nil {
		return fmt.Errorf("le dossier '%s' existe déjà", name)
	}

	// Créer le dossier
	if err := os.MkdirAll(name, 0755); err != nil {
		return fmt.Errorf("erreur création dossier: %v", err)
	}

	// Générer les fichiers selon le template
	generator := templates.NewGenerator(name, templateType, providers)

	files := []struct {
		name    string
		content string
	}{
		{"devly.json", generator.GenerateManifest()},
		{"index.js", generator.GenerateMainFile()},
		{"README.md", generator.GenerateReadme()},
		{".gitignore", generator.GenerateGitignore()},
	}

	for _, file := range files {
		filePath := filepath.Join(name, file.name)
		if err := os.WriteFile(filePath, []byte(file.content), 0644); err != nil {
			return fmt.Errorf("erreur création %s: %v", file.name, err)
		}
	}

	fmt.Printf("✅ Module '%s' créé avec succès!\n", name)
	fmt.Println()
	fmt.Println("📁 Structure générée:")
	fmt.Printf("   %s/\n", name)
	fmt.Println("   ├── devly.json")
	fmt.Println("   ├── index.js")
	fmt.Println("   ├── README.md")
	fmt.Println("   └── .gitignore")
	fmt.Println()
	fmt.Println("🔗 Prochaines étapes:")
	fmt.Printf("   cd %s\n", name)
	fmt.Println("   dashspace preview    # Tester dans Buildy")
	fmt.Println("   dashspace publish    # Publier sur le store")

	return nil
}
