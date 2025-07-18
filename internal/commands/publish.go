package commands

import (
	"fmt"
	"os"

	"github.com/devlyspace/devly-cli/internal/api"
	"github.com/devlyspace/devly-cli/internal/utils"
	"github.com/spf13/cobra"
)

func NewPublishCmd() *cobra.Command {
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "publish",
		Short: "Publish module to DashSpace store",
		RunE: func(cmd *cobra.Command, args []string) error {
			return publishModule(dryRun)
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Simulate without actual publishing")

	return cmd
}

func publishModule(dryRun bool) error {
	// Check if we're in a module directory
	if _, err := os.Stat("devly.json"); err != nil {
		return fmt.Errorf("❌ devly.json file not found")
	}

	fmt.Println("📦 Publishing module...")

	// Read manifest
	manifest, err := utils.ReadManifest("devly.json")
	if err != nil {
		return fmt.Errorf("error reading manifest: %v", err)
	}

	fmt.Printf("📋 Module: %s v%s\n", manifest.Name, manifest.Version)
	fmt.Printf("📝 Description: %s\n", manifest.Description)

	if dryRun {
		fmt.Println("🔍 Dry-run mode - no actual publishing")
		return nil
	}

	// Create ZIP archive
	fmt.Println("📁 Creating archive...")
	zipPath, err := utils.CreateModuleArchive(".")
	if err != nil {
		return fmt.Errorf("error creating archive: %v", err)
	}
	defer os.Remove(zipPath) // Clean up after

	fmt.Printf("📦 Archive created: %s\n", zipPath)

	// Publish via API
	client := api.NewClient()

	// Convert utils.Manifest to api.ModuleManifest
	apiManifest := &api.ModuleManifest{
		ID:          manifest.ID,
		Name:        manifest.Name,
		Version:     manifest.Version,
		Description: manifest.Description,
		Author:      manifest.Author,
		Providers:   manifest.Providers,
		Interfaces:  manifest.Interfaces,
	}

	// 1. Create or get module
	moduleID, err := client.CreateOrGetModule(apiManifest)
	if err != nil {
		return fmt.Errorf("error creating module: %v", err)
	}

	// 2. Upload new version
	fmt.Println("⬆️  Uploading...")
	versionID, err := client.UploadModuleVersion(moduleID, zipPath)
	if err != nil {
		return fmt.Errorf("error uploading: %v", err)
	}

	fmt.Printf("✅ Module published successfully!\n")
	fmt.Printf("🆔 Module ID: %d\n", moduleID)
	fmt.Printf("📦 Version ID: %d\n", versionID)
	fmt.Printf("🔗 Store: https://store.dashspace.dev/modules/%d\n", moduleID)

	return nil
}
