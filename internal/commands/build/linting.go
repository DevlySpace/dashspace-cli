package build

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strings"
)

type LintingValidator struct {
	projectPath string
}

func NewLintingValidator(projectPath string) *LintingValidator {
	return &LintingValidator{
		projectPath: projectPath,
	}
}

func (l *LintingValidator) RunESLint() error {
	// Check if ESLint is configured
	if err := l.ensureESLintConfig(); err != nil {
		return err
	}

	fmt.Println("🔍 Running ESLint checks...")

	cmd := exec.Command("npx", "eslint", ".", "--ext", ".ts,.tsx")
	cmd.Dir = l.projectPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if it's because ESLint isn't installed
		if strings.Contains(string(output), "not found") {
			fmt.Printf("⚠️  Warning: ESLint not installed. Run: npm install -D eslint @typescript-eslint/parser @typescript-eslint/eslint-plugin\n")
			return nil // Not fatal
		}

		if len(output) > 0 {
			fmt.Printf("❌ ESLint found issues:\n%s\n", output)
			return fmt.Errorf("ESLint validation failed")
		}
	}

	fmt.Println("✅ ESLint validation passed")
	return nil
}

func (l *LintingValidator) ensureESLintConfig() error {
	eslintConfigPath := filepath.Join(l.projectPath, ".eslintrc.json")

	if fileExists(eslintConfigPath) {
		return nil
	}

	fmt.Println("📝 Creating .eslintrc.json...")

	eslintConfig := map[string]interface{}{
		"parser": "@typescript-eslint/parser",
		"extends": []string{
			"eslint:recommended",
			"plugin:@typescript-eslint/recommended",
			"plugin:react/recommended",
			"plugin:react-hooks/recommended",
		},
		"plugins": []string{
			"@typescript-eslint",
			"react",
			"react-hooks",
		},
		"parserOptions": map[string]interface{}{
			"ecmaVersion": 2020,
			"sourceType":  "module",
			"ecmaFeatures": map[string]bool{
				"jsx": true,
			},
		},
		"settings": map[string]interface{}{
			"react": map[string]string{
				"version": "detect",
			},
		},
		"env": map[string]bool{
			"browser": true,
			"es2020":  true,
			"node":    true,
		},
		"rules": map[string]interface{}{
			"react/react-in-jsx-scope":                          "off",
			"@typescript-eslint/no-explicit-any":                "warn",
			"@typescript-eslint/explicit-module-boundary-types": "off",
			"no-console": []interface{}{"warn", map[string][]string{
				"allow": []string{"warn", "error"},
			}},
			"react/prop-types": "off", // Using TypeScript for prop validation
		},
		"ignorePatterns": []string{
			"dist/",
			"build/",
			"node_modules/",
			"*.js",
		},
	}

	jsonData, err := json.MarshalIndent(eslintConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to create .eslintrc.json: %w", err)
	}

	if err := ioutil.WriteFile(eslintConfigPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write .eslintrc.json: %w", err)
	}

	// Also create .eslintignore if it doesn't exist
	eslintIgnorePath := filepath.Join(l.projectPath, ".eslintignore")
	if !fileExists(eslintIgnorePath) {
		eslintIgnore := `dist/
build/
node_modules/
*.js
.interface-check.ts
`
		ioutil.WriteFile(eslintIgnorePath, []byte(eslintIgnore), 0644)
	}

	fmt.Println("✅ Created ESLint configuration")
	return nil
}

func (l *LintingValidator) CheckForCommonIssues() error {
	fmt.Println("🔍 Checking for common issues...")

	issues := []string{}

	// Check Module.ts
	moduleFile := findModuleFile()
	if moduleFile != "" {
		content, _ := ioutil.ReadFile(moduleFile)
		moduleContent := string(content)

		// Check for console.logs in production code
		if strings.Contains(moduleContent, "console.log") {
			issues = append(issues, "console.log found in Module.ts - consider removing for production")
		}

		// Check for hardcoded values that should be configurable
		if strings.Contains(moduleContent, "localhost") || strings.Contains(moduleContent, "http://127.0.0.1") {
			issues = append(issues, "Hardcoded localhost URL found - should be configurable")
		}

		// Check for TODO comments
		if strings.Contains(moduleContent, "TODO") || strings.Contains(moduleContent, "FIXME") {
			issues = append(issues, "TODO/FIXME comments found - resolve before publishing")
		}
	}

	// Check Component.tsx
	componentFile := findComponentFile()
	if componentFile != "" {
		content, _ := ioutil.ReadFile(componentFile)
		componentContent := string(content)

		// Check for any instead of proper types
		anyCount := strings.Count(componentContent, ": any")
		if anyCount > 0 {
			issues = append(issues, fmt.Sprintf("Found %d uses of 'any' type - consider using proper types", anyCount))
		}

		// Check for missing error boundaries
		if !strings.Contains(componentContent, "error") {
			issues = append(issues, "No error handling found in Component - consider adding error states")
		}

		// Check for missing loading states
		if !strings.Contains(componentContent, "loading") {
			issues = append(issues, "No loading state found in Component - consider adding loading indicators")
		}
	}

	if len(issues) > 0 {
		fmt.Println("⚠️  Found potential issues:")
		for _, issue := range issues {
			fmt.Printf("   - %s\n", issue)
		}
	} else {
		fmt.Println("✅ No common issues found")
	}

	return nil
}

func (l *LintingValidator) ValidatePackageJSON() error {
	fmt.Println("🔍 Validating package.json...")

	packageJsonPath := filepath.Join(l.projectPath, "package.json")
	packageData, err := ioutil.ReadFile(packageJsonPath)
	if err != nil {
		return fmt.Errorf("failed to read package.json: %w", err)
	}

	var packageJSON map[string]interface{}
	if err := json.Unmarshal(packageData, &packageJSON); err != nil {
		return fmt.Errorf("failed to parse package.json: %w", err)
	}

	issues := []string{}

	// Check required fields
	if _, ok := packageJSON["name"]; !ok {
		issues = append(issues, "Missing 'name' field")
	}

	if _, ok := packageJSON["version"]; !ok {
		issues = append(issues, "Missing 'version' field")
	}

	// Check scripts
	scripts, _ := packageJSON["scripts"].(map[string]interface{})
	if scripts == nil {
		issues = append(issues, "No scripts defined")
	} else {
		if _, ok := scripts["build"]; !ok {
			issues = append(issues, "Missing 'build' script")
		}
		if _, ok := scripts["test"]; !ok {
			fmt.Printf("⚠️  Warning: No 'test' script defined\n")
		}
	}

	// Check dependencies are properly categorized
	deps, _ := packageJSON["dependencies"].(map[string]interface{})
	devDeps, _ := packageJSON["devDependencies"].(map[string]interface{})

	// Check if dev tools are in dependencies instead of devDependencies
	if deps != nil {
		devTools := []string{"eslint", "typescript", "@types/react", "@types/react-dom", "prettier"}
		for _, tool := range devTools {
			if _, ok := deps[tool]; ok {
				issues = append(issues, fmt.Sprintf("'%s' should be in devDependencies, not dependencies", tool))
			}
		}
	}

	// Check if runtime deps are in devDependencies
	if devDeps != nil {
		runtimeDeps := []string{"react", "react-dom"}
		for _, dep := range runtimeDeps {
			if _, ok := devDeps[dep]; ok {
				issues = append(issues, fmt.Sprintf("'%s' should be in dependencies or peerDependencies, not devDependencies", dep))
			}
		}
	}

	if len(issues) > 0 {
		fmt.Println("❌ package.json validation issues:")
		for _, issue := range issues {
			fmt.Printf("   - %s\n", issue)
		}
		return fmt.Errorf("package.json validation failed")
	}

	fmt.Println("✅ package.json validation passed")
	return nil
}
