package commands

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/devlyspace/dashspace-cli/internal/api"
	"github.com/devlyspace/dashspace-cli/internal/config"
	"github.com/spf13/cobra"
)

func NewPublishCmd() *cobra.Command {
	var (
		dryRun   bool
		buildDir string
		signKey  string
	)

	cmd := &cobra.Command{
		Use:   "publish",
		Short: "Publish module to Dashspace store",
		RunE: func(cmd *cobra.Command, args []string) error {
			return publishModule(dryRun, buildDir, signKey)
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Simulate without actual publishing")
	cmd.Flags().StringVar(&buildDir, "build-dir", "dist", "Build directory containing the module")
	cmd.Flags().StringVarP(&signKey, "key", "k", "", "Signature key for module signing")

	return cmd
}

func publishModule(dryRun bool, buildDir string, signKey string) error {
	fmt.Println("📦 Publishing module to Dashspace...")

	// 1. Check if build directory exists
	if _, err := os.Stat(buildDir); err != nil {
		fmt.Println("❌ Build directory not found. Running build first...")

		// Run build command
		buildCmd := NewBuildCmd()
		buildCmd.SetArgs([]string{"--output", buildDir})
		if err := buildCmd.Execute(); err != nil {
			return fmt.Errorf("build failed: %v", err)
		}
	}

	// 2. Check for dashspace.json
	dashspacePath := filepath.Join(buildDir, "dashspace.json")
	if _, err := os.Stat(dashspacePath); err != nil {
		fmt.Println("⚠️  dashspace.json not found. Building module...")

		// Run build to generate dashspace.json
		buildCmd := NewBuildCmd()
		buildCmd.SetArgs([]string{"--output", buildDir, "--key", signKey})
		if err := buildCmd.Execute(); err != nil {
			return fmt.Errorf("build failed: %v", err)
		}
	}

	// 3. Read dashspace.json
	dashspaceData, err := os.ReadFile(dashspacePath)
	if err != nil {
		return fmt.Errorf("failed to read dashspace.json: %v", err)
	}

	var dashspaceConfig DashspaceConfig
	if err := json.Unmarshal(dashspaceData, &dashspaceConfig); err != nil {
		return fmt.Errorf("invalid dashspace.json: %v", err)
	}

	fmt.Printf("📋 Module: %s v%s\n", dashspaceConfig.Name, dashspaceConfig.Version)
	fmt.Printf("📝 Description: %s\n", dashspaceConfig.Description)
	fmt.Printf("👤 Author: %s\n", dashspaceConfig.Author)

	if dryRun {
		fmt.Println("🔍 Dry-run mode - no actual publishing")
		fmt.Println("\nFiles to be published:")

		// List files that would be included
		err := filepath.Walk(buildDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			relPath, _ := filepath.Rel(buildDir, path)
			if !info.IsDir() && relPath != "" {
				fmt.Printf("  - %s (%d bytes)\n", relPath, info.Size())
			}
			return nil
		})

		return err
	}

	// 4. Create ZIP archive
	fmt.Println("📦 Creating archive...")
	zipPath, err := createModuleZip(buildDir, dashspaceConfig)
	if err != nil {
		return fmt.Errorf("failed to create archive: %v", err)
	}
	defer os.Remove(zipPath)

	fileInfo, _ := os.Stat(zipPath)
	fmt.Printf("📦 Archive created: %s (%.2f KB)\n", filepath.Base(zipPath), float64(fileInfo.Size())/1024)

	// 5. Check authentication
	cfg := config.GetConfig()
	if cfg.AuthToken == "" {
		return fmt.Errorf("❌ Not logged in. Run 'dashspace login' first")
	}

	// 6. Publish via API
	client := api.NewClient()

	// Convert to API manifest
	apiManifest := &api.ModuleManifest{
		ID:          dashspaceConfig.Name,
		Name:        dashspaceConfig.Name,
		Version:     dashspaceConfig.Version,
		Description: dashspaceConfig.Description,
		Author:      dashspaceConfig.Author,
	}

	// Create or get module
	fmt.Println("🔍 Checking module existence...")
	moduleID, err := client.CreateOrGetModule(apiManifest)
	if err != nil {
		return fmt.Errorf("failed to create/get module: %v", err)
	}

	// Upload new version
	fmt.Println("⬆️  Uploading module...")
	versionID, err := client.UploadModuleVersion(moduleID, zipPath)
	if err != nil {
		return fmt.Errorf("upload failed: %v", err)
	}

	fmt.Printf("\n✅ Module published successfully!\n")
	fmt.Printf("🆔 Module ID: %d\n", moduleID)
	fmt.Printf("📦 Version ID: %d\n", versionID)
	fmt.Printf("🔗 Store: https://store.dashspace.dev/modules/%d\n", moduleID)

	return nil
}

func createModuleZip(buildDir string, config DashspaceConfig) (string, error) {
	// Create temporary zip file
	zipPath := filepath.Join(os.TempDir(), fmt.Sprintf("%s-%s.zip", config.Name, config.Version))

	zipFile, err := os.Create(zipPath)
	if err != nil {
		return "", err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Walk through build directory and add files
	err = filepath.Walk(buildDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path
		relPath, err := filepath.Rel(buildDir, path)
		if err != nil {
			return err
		}

		// Skip the build directory itself
		if relPath == "." {
			return nil
		}

		// Create zip entry
		if info.IsDir() {
			// Add directory
			_, err := zipWriter.Create(relPath + "/")
			return err
		}

		// Add file
		zipFileWriter, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}

		// Open and copy file
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(zipFileWriter, file)
		return err
	})

	if err != nil {
		os.Remove(zipPath)
		return "", err
	}

	return zipPath, nil
}