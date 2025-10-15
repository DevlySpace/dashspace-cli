package commands

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/devlyspace/dashspace-cli/internal/commands/build"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func NewPublishCmd() *cobra.Command {
	var (
		dryRun   bool
		buildDir string
		signKey  string
		moduleID int
		modlyURL string
	)

	cmd := &cobra.Command{
		Use:   "publish",
		Short: "Publish module to Dashspace store",
		Long:  "Build and publish a module to the Dashspace store",
		RunE: func(cmd *cobra.Command, args []string) error {
			return publishModule(dryRun, buildDir, signKey, moduleID, modlyURL)
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Simulate without actual publishing")
	cmd.Flags().StringVar(&buildDir, "build-dir", "dist", "Build directory containing the module")
	cmd.Flags().StringVarP(&signKey, "key", "k", "", "Signature key for module signing")
	cmd.Flags().IntVarP(&moduleID, "module-id", "m", 0, "Override module ID (uses ID from dashspace.json by default)")
	cmd.Flags().StringVar(&modlyURL, "url", "http://localhost:10001", "Modly API URL")

	return cmd
}

func publishModule(dryRun bool, buildDir string, signKey string, overrideModuleID int, modlyURL string) error {
	fmt.Println("📦 Publishing module to Dashspace...")

	if _, err := os.Stat(buildDir); err != nil {
		fmt.Println("❌ Build directory not found. Running build first...")
		buildCmd := build.NewBuildCmd()
		args := []string{"--output", buildDir}
		if signKey != "" {
			args = append(args, "--key", signKey)
		}
		buildCmd.SetArgs(args)
		if err := buildCmd.Execute(); err != nil {
			return fmt.Errorf("build failed: %v", err)
		}
	}

	dashspacePath := filepath.Join(buildDir, "dashspace.json")
	if _, err := os.Stat(dashspacePath); err != nil {
		return fmt.Errorf("dashspace.json not found in %s. Run 'dashspace build' first", buildDir)
	}

	dashspaceData, err := os.ReadFile(dashspacePath)
	if err != nil {
		return fmt.Errorf("failed to read dashspace.json: %v", err)
	}

	var dashspaceConfig map[string]interface{}
	if err := json.Unmarshal(dashspaceData, &dashspaceConfig); err != nil {
		return fmt.Errorf("invalid dashspace.json: %v", err)
	}

	moduleID := int(dashspaceConfig["id"].(float64))
	if overrideModuleID != 0 {
		moduleID = overrideModuleID
		dashspaceConfig["id"] = moduleID
		fmt.Printf("⚠️  Using override module ID: %d\n", moduleID)

		updatedJSON, _ := json.MarshalIndent(dashspaceConfig, "", "  ")
		if err := os.WriteFile(dashspacePath, updatedJSON, 0644); err != nil {
			return fmt.Errorf("failed to update dashspace.json with override ID: %v", err)
		}
	}

	if moduleID == 0 {
		return fmt.Errorf("module ID not found. Set it in Module.ts or use -m flag")
	}

	fmt.Printf("🆔 Module ID: %d\n", moduleID)
	fmt.Printf("📛 Module Slug: %s\n", dashspaceConfig["slug"])
	fmt.Printf("📋 Module: %s v%s\n", dashspaceConfig["name"], dashspaceConfig["version"])
	fmt.Printf("📝 Description: %s\n", dashspaceConfig["description"])
	fmt.Printf("👤 Author: %s\n", dashspaceConfig["author"])

	if requiresSetup, ok := dashspaceConfig["requires_setup"].(bool); ok && requiresSetup {
		fmt.Printf("⚙️  Requires Setup: Yes\n")
		if configSteps, ok := dashspaceConfig["configuration_steps"].([]interface{}); ok {
			fmt.Printf("📋 Configuration Steps: %d\n", len(configSteps))
		}
	}

	if dryRun {
		fmt.Println("\n🔍 Dry-run mode - no actual publishing")
		fmt.Println("Files to be published:")
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
		fmt.Printf("\n📡 Would upload to: %s/modules/%d/module_versions/upload\n", modlyURL, moduleID)

		if requiresSetup, ok := dashspaceConfig["requires_setup"].(bool); ok && requiresSetup {
			fmt.Println("\n📋 Configuration that would be sent:")
			if configSteps, ok := dashspaceConfig["configuration_steps"]; ok {
				configJSON, _ := json.MarshalIndent(configSteps, "  ", "  ")
				fmt.Printf("  %s\n", string(configJSON))
			}
		}

		return err
	}

	fmt.Println("📦 Creating archive...")
	zipPath, err := createModuleZip(buildDir, dashspaceConfig)
	if err != nil {
		return fmt.Errorf("failed to create archive: %v", err)
	}
	defer os.Remove(zipPath)

	fileInfo, _ := os.Stat(zipPath)
	fmt.Printf("📦 Archive created: %s (%.2f KB)\n", filepath.Base(zipPath), float64(fileInfo.Size())/1024)

	fmt.Printf("⬆️  Uploading to %s...\n", modlyURL)

	responseData, err := uploadToModly(modlyURL, moduleID, zipPath, dashspaceConfig)
	if err != nil {
		return fmt.Errorf("upload failed: %v", err)
	}

	fmt.Printf("\n✅ Module published successfully!\n")
	fmt.Printf("🆔 Module ID: %d\n", moduleID)
	fmt.Printf("📛 Module Slug: %s\n", dashspaceConfig["slug"])
	fmt.Printf("📦 Version: %s\n", dashspaceConfig["version"])

	if versionId, ok := responseData["id"]; ok {
		fmt.Printf("📦 Version ID: %v\n", versionId)
	}

	if requiresSetup, ok := responseData["requires_setup"].(bool); ok && requiresSetup {
		fmt.Printf("⚙️  Setup Required: Yes\n")
	}

	fmt.Printf("🔗 API: %s/modules/%d\n", modlyURL, moduleID)

	return nil
}

func createModuleZip(buildDir string, config map[string]interface{}) (string, error) {
	slug := config["slug"].(string)
	version := config["version"].(string)
	zipPath := filepath.Join(os.TempDir(), fmt.Sprintf("%s-%s.zip", slug, version))

	zipFile, err := os.Create(zipPath)
	if err != nil {
		return "", err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	err = filepath.Walk(buildDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(buildDir, path)
		if err != nil {
			return err
		}

		if relPath == "." {
			return nil
		}

		if info.IsDir() {
			_, err := zipWriter.Create(relPath + "/")
			return err
		}

		zipFileWriter, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}

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

func uploadToModly(modlyURL string, moduleID int, zipPath string, dashspaceConfig map[string]interface{}) (map[string]interface{}, error) {
	file, err := os.Open(zipPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open zip file: %v", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(zipPath))
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %v", err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return nil, fmt.Errorf("failed to copy file: %v", err)
	}

	if metadata, err := json.Marshal(dashspaceConfig); err == nil {
		if err := writer.WriteField("metadata", string(metadata)); err != nil {
			fmt.Printf("⚠️  Warning: Failed to add metadata field: %v\n", err)
		}
	}

	if requiresSetup, ok := dashspaceConfig["requires_setup"].(bool); ok && requiresSetup {
		if err := writer.WriteField("requires_setup", "true"); err != nil {
			fmt.Printf("⚠️  Warning: Failed to add requires_setup field: %v\n", err)
		}

		if configSteps, ok := dashspaceConfig["configuration_steps"]; ok {
			if configJSON, err := json.Marshal(configSteps); err == nil {
				if err := writer.WriteField("configuration_steps", string(configJSON)); err != nil {
					fmt.Printf("⚠️  Warning: Failed to add configuration_steps field: %v\n", err)
				}
			}
		}
	}

	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close writer: %v", err)
	}

	url := fmt.Sprintf("%s/modules/%d/module_versions/upload", modlyURL, moduleID)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	responseBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	fmt.Printf("📡 Server response: %s\n", resp.Status)

	var result map[string]interface{}
	if len(responseBody) > 0 {
		if err := json.Unmarshal(responseBody, &result); err == nil {
			return result, nil
		}
	}

	return map[string]interface{}{}, nil
}
