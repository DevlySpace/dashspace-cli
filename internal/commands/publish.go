package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/devlyspace/devly-cli/internal/api"
	"github.com/devlyspace/devly-cli/internal/utils"
	"github.com/spf13/cobra"
)

func NewPublishCmd() *cobra.Command {
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "publish",
		Short: "Publier le module sur le store DashSpace",
		RunE: func(cmd *cobra.Command, args []string) error {
			return publishModule(dryRun)
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Simulation sans publication réelle")

	return cmd
}

func publishModule(dryRun bool) error {
	// Vérifier qu'on est dans un dossier de module
	if _, err := os.Stat("devly.json"); err != nil {
		return fmt.Errorf("❌ Pas de fichier devly.json trouvé")
	}

	fmt.Println("📦 Publication du module...")

	// Lire le manifest
	manifest, err := utils.ReadManifest("devly.json")
	if err != nil {
		return fmt.Errorf("erreur lecture manifest: %v", err)
	}

	fmt.Printf("📋 Module: %s v%s\n", manifest.Name, manifest.Version)
	fmt.Printf("📝 Description: %s\n", manifest.Description)

	if dryRun {
		fmt.Println("🔍 Mode dry-run - aucune publication réelle")
		return nil
	}

	// Créer l'archive ZIP
	fmt.Println("📁 Création de l'archive...")
	zipPath, err := utils.CreateModuleArchive(".")
	if err != nil {
		return fmt.Errorf("erreur création archive: %v", err)
	}
	defer os.Remove(zipPath) // Nettoyer après

	fmt.Printf("📦 Archive créée: %s\n", zipPath)

	// Publier via l'API
	client := api.NewClient()

	// 1. Créer ou récupérer le module
	moduleID, err := client.CreateOrGetModule(manifest)
	if err != nil {
		return fmt.Errorf("erreur création module: %v", err)
	}

	// 2. Upload la nouvelle version
	fmt.Println("⬆️  Upload en cours...")
	versionID, err := client.UploadModuleVersion(moduleID, zipPath)
	if err != nil {
		return fmt.Errorf("erreur upload: %v", err)
	}

	fmt.Printf("✅ Module publié avec succès!\n")
	fmt.Printf("🆔 Module ID: %d\n", moduleID)
	fmt.Printf("📦 Version ID: %d\n", versionID)
	fmt.Printf("🔗 Store: https://store.dashspace.dev/modules/%d\n", moduleID)

	return nil
}
