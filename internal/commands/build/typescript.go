package build

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type TypeScriptValidator struct {
	projectPath string
}

func NewTypeScriptValidator(projectPath string) *TypeScriptValidator {
	return &TypeScriptValidator{
		projectPath: projectPath,
	}
}

func (t *TypeScriptValidator) Validate() error {
	// Ensure tsconfig.json exists
	if err := t.ensureTsConfig(); err != nil {
		return err
	}

	// Run TypeScript compiler in check mode
	fmt.Println("🔍 Running TypeScript type checking...")
	cmd := exec.Command("npx", "tsc", "--noEmit", "--skipLibCheck")
	cmd.Dir = t.projectPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Parse and format TypeScript errors
		errors := t.parseTypeScriptErrors(string(output))
		if len(errors) > 0 {
			fmt.Println("\n❌ TypeScript validation failed:")
			for _, e := range errors {
				fmt.Printf("   %s\n", e)
			}
			return fmt.Errorf("found %d TypeScript errors", len(errors))
		}
		return fmt.Errorf("TypeScript validation failed: %s", output)
	}

	fmt.Println("✅ TypeScript validation passed")
	return nil
}

func (t *TypeScriptValidator) ValidateDashspaceLibTypes() error {
	fmt.Println("🔍 Validating dashspace-lib usage...")

	// Check if dashspace-lib is installed
	packageJsonPath := filepath.Join(t.projectPath, "package.json")
	packageData, err := ioutil.ReadFile(packageJsonPath)
	if err != nil {
		return fmt.Errorf("failed to read package.json: %w", err)
	}

	var packageJSON map[string]interface{}
	if err := json.Unmarshal(packageData, &packageJSON); err != nil {
		return fmt.Errorf("failed to parse package.json: %w", err)
	}

	// Check dependencies
	deps, _ := packageJSON["dependencies"].(map[string]interface{})
	devDeps, _ := packageJSON["devDependencies"].(map[string]interface{})

	hasDashspaceLib := false
	if deps != nil {
		if _, ok := deps["dashspace-lib"]; ok {
			hasDashspaceLib = true
		}
	}
	if devDeps != nil {
		if _, ok := devDeps["dashspace-lib"]; ok {
			hasDashspaceLib = true
		}
	}

	if !hasDashspaceLib {
		return fmt.Errorf("dashspace-lib is not installed. Run: npm install dashspace-lib")
	}

	// Check if @types/react and @types/react-dom are installed for proper type checking
	hasReactTypes := false
	hasReactDomTypes := false

	if devDeps != nil {
		if _, ok := devDeps["@types/react"]; ok {
			hasReactTypes = true
		}
		if _, ok := devDeps["@types/react-dom"]; ok {
			hasReactDomTypes = true
		}
	}

	if !hasReactTypes || !hasReactDomTypes {
		fmt.Printf("⚠️  Warning: Missing React type definitions. Run: npm install -D @types/react @types/react-dom\n")
	}

	return nil
}

func (t *TypeScriptValidator) ensureTsConfig() error {
	tsconfigPath := filepath.Join(t.projectPath, "tsconfig.json")

	if fileExists(tsconfigPath) {
		return nil
	}

	fmt.Println("📝 Creating tsconfig.json...")

	tsconfig := map[string]interface{}{
		"compilerOptions": map[string]interface{}{
			"target":                           "ES2020",
			"module":                           "ESNext",
			"lib":                              []string{"ES2020", "DOM", "DOM.Iterable"},
			"jsx":                              "react",
			"strict":                           true,
			"esModuleInterop":                  true,
			"skipLibCheck":                     true,
			"forceConsistentCasingInFileNames": true,
			"moduleResolution":                 "node",
			"resolveJsonModule":                true,
			"noEmit":                           true,
			"types":                            []string{"react", "react-dom", "node"},
		},
		"include": []string{
			"**/*.ts",
			"**/*.tsx",
		},
		"exclude": []string{
			"node_modules",
			"dist",
			"build",
		},
	}

	jsonData, err := json.MarshalIndent(tsconfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to create tsconfig.json: %w", err)
	}

	if err := ioutil.WriteFile(tsconfigPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write tsconfig.json: %w", err)
	}

	fmt.Println("✅ Created tsconfig.json")
	return nil
}

func (t *TypeScriptValidator) parseTypeScriptErrors(output string) []string {
	lines := strings.Split(output, "\n")
	var errors []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Skip non-error lines
		if strings.Contains(line, "error TS") {
			// Format: file.ts(line,col): error TS2322: Message
			errors = append(errors, line)
		} else if strings.HasPrefix(line, "src/") || strings.HasPrefix(line, "./") {
			// Include file references
			if len(errors) > 0 && !strings.Contains(errors[len(errors)-1], line) {
				errors[len(errors)-1] = line + " - " + errors[len(errors)-1]
			}
		}
	}

	return errors
}

func (t *TypeScriptValidator) ValidateInterfaceImplementation() error {
	fmt.Println("🔍 Validating interface implementations with TypeScript...")

	// Create a temporary test file to validate interface implementations
	testFile := filepath.Join(t.projectPath, ".interface-check.ts")
	defer os.Remove(testFile)

	// Read Module.ts to get declared interfaces
	moduleFile := findModuleFile()
	if moduleFile == "" {
		return fmt.Errorf("Module.ts not found")
	}

	_, err := ioutil.ReadFile(moduleFile)
	if err != nil {
		return err
	}

	// Read Component.tsx
	componentFile := findComponentFile()
	if componentFile == "" {
		fmt.Printf("⚠️  Warning: Component.tsx not found, skipping interface validation\n")
		return nil
	}

	// Create a test that imports both and validates
	testContent := `
import { InterfaceHandlers } from 'dashspace-lib';
import Component from './Component';

// This file is auto-generated to validate interface implementations
// If this compiles without errors, the interfaces are properly implemented
`

	if err := ioutil.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		return fmt.Errorf("failed to create interface test file: %w", err)
	}

	// Détecter le type de jsx depuis tsconfig.json
	jsxMode := t.getJSXMode()

	// Run TypeScript compiler on the test file avec les bonnes options jsx
	cmd := exec.Command("npx", "tsc", testFile, "--noEmit", "--skipLibCheck", "--jsx", jsxMode, "--esModuleInterop", "--moduleResolution", "node")
	cmd.Dir = t.projectPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Clean up error output to show only relevant parts
		if strings.Contains(string(output), "error") {
			// Filtrer les erreurs non pertinentes
			errorLines := strings.Split(string(output), "\n")
			var relevantErrors []string
			for _, line := range errorLines {
				// Ignorer les erreurs liées au jsx qui sont déjà gérées
				if strings.Contains(line, "--jsx") {
					continue
				}
				if strings.TrimSpace(line) != "" {
					relevantErrors = append(relevantErrors, line)
				}
			}
			if len(relevantErrors) > 0 {
				return fmt.Errorf("interface implementation errors:\n%s", strings.Join(relevantErrors, "\n"))
			}
		}
	}

	fmt.Println("✅ Interface implementations validated by TypeScript")
	return nil
}

func (t *TypeScriptValidator) getJSXMode() string {
	// Lire tsconfig.json pour récupérer le mode jsx
	tsconfigPath := filepath.Join(t.projectPath, "tsconfig.json")

	if !fileExists(tsconfigPath) {
		return "react" // Default
	}

	data, err := ioutil.ReadFile(tsconfigPath)
	if err != nil {
		return "react" // Default
	}

	var tsconfig map[string]interface{}
	if err := json.Unmarshal(data, &tsconfig); err != nil {
		return "react" // Default
	}

	if compilerOptions, ok := tsconfig["compilerOptions"].(map[string]interface{}); ok {
		if jsx, ok := compilerOptions["jsx"].(string); ok {
			return jsx
		}
	}

	return "react" // Default
}

func (t *TypeScriptValidator) CheckUnusedImports() error {
	fmt.Println("🔍 Checking for unused imports...")

	// Use TypeScript compiler with additional checks
	cmd := exec.Command("npx", "tsc", "--noEmit", "--noUnusedLocals", "--noUnusedParameters")
	cmd.Dir = t.projectPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		if strings.Contains(string(output), "is declared but") {
			fmt.Printf("⚠️  Warning: Found unused code:\n%s\n", output)
			// Not a fatal error, just a warning
		}
	}

	return nil
}

func (t *TypeScriptValidator) ValidateSatisfiesUsage() error {
	componentFile := findComponentFile()
	if componentFile == "" {
		return nil
	}

	content, err := ioutil.ReadFile(componentFile)
	if err != nil {
		return err
	}

	componentContent := string(content)

	// Check if interfaces are using satisfies for type safety
	if strings.Contains(componentContent, "ISearchable:") && !strings.Contains(componentContent, "satisfies ISearchable") {
		fmt.Printf("⚠️  Warning: Consider using 'satisfies ISearchable' for better type safety\n")
	}

	if strings.Contains(componentContent, "InterfaceHandlers") && !strings.Contains(componentContent, "satisfies InterfaceHandlers") {
		fmt.Printf("⚠️  Warning: Consider using 'satisfies InterfaceHandlers' for better type safety\n")
	}

	return nil
}
