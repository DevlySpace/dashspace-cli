package commands

import (
	"archive/zip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

type DashspaceConfig struct {
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Description string                 `json:"description"`
	Author      string                 `json:"author"`
	Entry       string                 `json:"entry"`
	Icon        string                 `json:"icon,omitempty"`
	Category    string                 `json:"category,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Providers   []ProviderDefinition  `json:"providers,omitempty"`
	Config      []ConfigStep           `json:"config"`
	Security    SecurityInfo           `json:"security"`
}

type ProviderDefinition struct {
	Name        string   `json:"name"`
	Required    bool     `json:"required"`
	Scopes      []string `json:"scopes,omitempty"`
	Description string   `json:"description,omitempty"`
}

type ConfigStep struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description,omitempty"`
	Order       int                    `json:"order"`
	Optional    bool                   `json:"optional"`
	Fields      map[string]interface{} `json:"fields"`
}

type SecurityInfo struct {
	Signature string `json:"signature"`
	Checksum  string `json:"checksum"`
	Algorithm string `json:"algorithm"`
	Timestamp string `json:"timestamp"`
}

func NewBuildCmd() *cobra.Command {
	var (
		output   string
		minify   bool
		noTS     bool
		signKey  string
	)

	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build a Dashspace module",
		Long:  "Compile TypeScript/React code and create a module package",
		RunE: func(cmd *cobra.Command, args []string) error {
			return buildModule(output, minify, noTS, signKey)
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "dist", "Output directory")
	cmd.Flags().BoolVar(&minify, "minify", false, "Minify the output")
	cmd.Flags().BoolVar(&noTS, "no-typescript", false, "Skip TypeScript compilation")
	cmd.Flags().StringVarP(&signKey, "key", "k", "", "Signature key (uses env DASHSPACE_KEY if not set)")

	return cmd
}

func buildModule(outputDir string, minify, noTS bool, signKey string) error {
	fmt.Println("🔨 Building Dashspace module...")

	// 1. Check for package.json
	if _, err := os.Stat("package.json"); err != nil {
		return fmt.Errorf("❌ package.json not found. Are you in a module directory?")
	}

	// 2. Install dependencies if needed
	if _, err := os.Stat("node_modules"); err != nil {
		fmt.Println("📦 Installing dependencies...")
		cmd := exec.Command("npm", "install")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install dependencies: %v", err)
		}
	}

	// 3. Compile TypeScript if needed
	if !noTS {
		fmt.Println("📦 Compiling TypeScript...")
		if err := compileTypeScript(); err != nil {
			return fmt.Errorf("TypeScript compilation failed: %v", err)
		}
	}

	// 4. Extract module metadata from compiled code
	fmt.Println("📋 Extracting module metadata...")
	config, err := extractModuleConfig()
	if err != nil {
		return fmt.Errorf("failed to extract config: %v", err)
	}

	// 5. Generate dashspace.json
	fmt.Println("📝 Generating dashspace.json...")
	dashspaceJSON, err := generateDashspaceJSON(config, signKey)
	if err != nil {
		return fmt.Errorf("failed to generate dashspace.json: %v", err)
	}

	// 6. Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// 7. Write dashspace.json
	dashspacePath := filepath.Join(outputDir, "dashspace.json")
	if err := os.WriteFile(dashspacePath, dashspaceJSON, 0644); err != nil {
		return fmt.Errorf("failed to write dashspace.json: %v", err)
	}

	// 8. Copy compiled files to output
	if err := copyCompiledFiles(outputDir, minify); err != nil {
		return fmt.Errorf("failed to copy compiled files: %v", err)
	}

	// 9. Display success info
	fmt.Printf("\n✅ Module built successfully!\n")
	fmt.Printf("📦 Output: %s/\n", outputDir)
	fmt.Printf("📋 Config: %s/dashspace.json\n", outputDir)
	fmt.Printf("🔑 Signed: %v\n", signKey != "")

	return nil
}

func compileTypeScript() error {
	// Check if TypeScript is installed
	if _, err := exec.LookPath("tsc"); err != nil {
		// Try with npx
		cmd := exec.Command("npx", "tsc")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	cmd := exec.Command("tsc")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func extractModuleConfig() (*DashspaceConfig, error) {
	// Create a Node.js script to extract config
	extractScript := `
const path = require('path');

// Try to load the compiled module
let Module;
try {
    Module = require('./dist/index.js');
} catch (err) {
    console.error(JSON.stringify({error: 'Failed to load module: ' + err.message}));
    process.exit(1);
}

// Find the module class
let ModuleClass = null;
for (const key in Module) {
    if (Module[key] && Module[key].prototype) {
        const proto = Module[key].prototype;
        // Check if it has the required methods
        if (proto.defineConfigurationSteps && proto.initialize && proto.callProvider) {
            ModuleClass = Module[key];
            break;
        }
    }
}

if (!ModuleClass) {
    // Try default export
    if (Module.default && Module.default.prototype) {
        ModuleClass = Module.default;
    }
}

if (!ModuleClass) {
    console.error(JSON.stringify({error: 'No Dashspace module class found'}));
    process.exit(1);
}

try {
    const instance = new ModuleClass();
    const metadata = instance.getMetadata();
    const steps = instance.getConfigurationSteps();
    const providers = instance.getProviders ? instance.getProviders() : instance.defineProviders ? instance.defineProviders() : [];

    const config = {
        name: metadata.id,
        version: metadata.version,
        description: metadata.description,
        author: metadata.author,
        entry: 'index.js',
        icon: metadata.icon,
        category: metadata.category,
        tags: metadata.tags,
        providers: providers,
        config: steps.map(step => ({
            id: step.id,
            title: step.title,
            description: step.description,
            order: step.order,
            optional: step.optional || false,
            fields: step.fields.reduce((acc, field) => {
                const fieldConfig = {
                    type: field.type,
                    title: field.label,
                    description: field.description,
                    required: field.validation?.required || false,
                    default: field.defaultValue,
                    placeholder: field.placeholder
                };

                // Add type-specific properties
                if (field.type === 'number') {
                    if (field.min !== undefined) fieldConfig.min = field.min;
                    if (field.max !== undefined) fieldConfig.max = field.max;
                    if (field.step !== undefined) fieldConfig.step = field.step;
                }

                if ((field.type === 'select' || field.type === 'multiselect') && field.options) {
                    fieldConfig.options = field.options.map(opt => opt.value);
                    if (field.multiple) fieldConfig.multiple = true;
                }

                if (field.validation?.pattern) {
                    fieldConfig.pattern = field.validation.pattern;
                }

                if (field.minLength !== undefined) fieldConfig.minLength = field.minLength;
                if (field.maxLength !== undefined) fieldConfig.maxLength = field.maxLength;

                acc[field.name] = fieldConfig;
                return acc;
            }, {})
        }))
    };

    console.log(JSON.stringify(config));
} catch (err) {
    console.error(JSON.stringify({error: 'Failed to extract config: ' + err.message}));
    process.exit(1);
}
`

	// Write temporary script
	tmpFile := "extract-config.js"
	if err := os.WriteFile(tmpFile, []byte(extractScript), 0644); err != nil {
		return nil, err
	}
	defer os.Remove(tmpFile)

	// Execute script
	cmd := exec.Command("node", tmpFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to extract config: %s", string(output))
	}

	// Parse JSON response
	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse config: %v", err)
	}

	// Check for error
	if errMsg, ok := result["error"].(string); ok {
		return nil, fmt.Errorf(errMsg)
	}

	// Convert to DashspaceConfig
	configBytes, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	var config DashspaceConfig
	if err := json.Unmarshal(configBytes, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func generateDashspaceJSON(config *DashspaceConfig, signKey string) ([]byte, error) {
	// Use environment key if not provided
	if signKey == "" {
		signKey = os.Getenv("DASHSPACE_KEY")
	}

	// Generate checksum
	configBytes, _ := json.Marshal(config)
	hash := sha256.Sum256(configBytes)
	checksum := hex.EncodeToString(hash[:])

	// Add security info
	config.Security = SecurityInfo{
		Checksum:  checksum,
		Algorithm: "sha256",
		Timestamp: time.Now().Format(time.RFC3339),
	}

	// Generate signature if key provided
	if signKey != "" {
		h := hmac.New(sha256.New, []byte(signKey))
		h.Write(configBytes)
		config.Security.Signature = hex.EncodeToString(h.Sum(nil))
	}

	return json.MarshalIndent(config, "", "  ")
}

func copyCompiledFiles(outputDir string, minify bool) error {
	// Check if dist directory exists
	if _, err := os.Stat("dist"); err != nil {
		return fmt.Errorf("dist directory not found. Run TypeScript compilation first")
	}

	// Copy all files from dist to output directory
	return filepath.Walk("dist", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the dist directory itself
		if path == "dist" {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel("dist", path)
		if err != nil {
			return err
		}

		// Destination path
		destPath := filepath.Join(outputDir, relPath)

		// Create directories
		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}

		// Copy file
		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		destFile, err := os.Create(destPath)
		if err != nil {
			return err
		}
		defer destFile.Close()

		_, err = io.Copy(destFile, srcFile)
		return err
	})
}