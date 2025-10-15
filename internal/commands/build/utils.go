package build

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func findEntryPoint() string {
	candidates := []string{
		"Module.tsx",
		"Module.ts",
		"src/Module.tsx",
		"src/Module.ts",
		"index.tsx",
		"index.ts",
		"src/index.tsx",
		"src/index.ts",
	}

	for _, candidate := range candidates {
		if fileExists(candidate) {
			return candidate
		}
	}
	return ""
}

func findModuleFile() string {
	candidates := []string{
		"Module.tsx",
		"Module.ts",
		"src/Module.tsx",
		"src/Module.ts",
	}

	for _, candidate := range candidates {
		if fileExists(candidate) {
			return candidate
		}
	}
	return ""
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func sanitizeSlug(name string) string {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "@", "")
	slug = strings.ReplaceAll(slug, "/", "-")
	slug = strings.ReplaceAll(slug, "_", "-")
	return slug
}

func ensureDependencies() error {
	if _, err := os.Stat("node_modules"); err != nil {
		fmt.Println("📦 Installing dependencies...")
		cmd := exec.Command("npm", "install")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install dependencies: %w", err)
		}
	}
	return nil
}
