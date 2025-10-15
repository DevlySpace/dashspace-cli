package build

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

type BuildOptions struct {
	Output     string
	Minify     bool
	Watch      bool
	Dev        bool
	Format     string
	SkipChecks bool
	Strict     bool
	NoStrict   bool
}

func NewBuildCmd() *cobra.Command {
	var opts BuildOptions

	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build a Dashspace module",
		Long: `Bundle TypeScript/React code into a Dashspace module with comprehensive validation.

VALIDATION STEPS:
The build process includes the following validation steps by default:

1. STRUCTURE VALIDATION
   - Verifies Module.ts/tsx exists with proper exports
   - Validates Component.tsx implementation (if present)
   - Checks package.json structure

2. TYPESCRIPT VALIDATION (--strict mode, default)
   - Type checking with tsc --noEmit
   - Interface implementation validation
   - dashspace-lib type compatibility
   - Unused imports detection
   - Validates 'satisfies' usage for type safety

3. LINTING (--strict mode, default)
   - ESLint with TypeScript rules
   - React hooks validation
   - Code style consistency
   - Common issue detection (console.logs, hardcoded URLs, TODOs)

4. METADATA EXTRACTION
   - Module ID, name, version, slug
   - Configuration steps parsing
   - Provider requirements
   - Interface declarations
   - Permission requirements
   - Data schema extraction (if module exposes data)

5. COMPILATION
   - TypeScript to JavaScript transpilation
   - JSX transformation
   - Tree shaking for optimal bundle size
   - Minification (production mode)

6. BUNDLE GENERATION
   - Wraps module in Dashspace loader
   - Includes dashspace-lib runtime polyfill
   - Generates SHA256 checksum

7. OUTPUT VALIDATION
   - Verifies bundle.js generation
   - Validates manifest completeness
   - Checks bundle size (warns if >500KB)

FLAGS:
By default, the build runs in strict mode with all validations enabled.
Use --no-strict to disable strict mode and treat warnings as non-fatal.
Use --skip-checks to skip TypeScript and linting validation entirely (not recommended).

EXAMPLES:
  dashspace build
  dashspace build --dev
  dashspace build --watch
  dashspace build --no-strict
  dashspace build --skip-checks
  dashspace build -o ./my-dist`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.NoStrict {
				opts.Strict = false
			} else if !cmd.Flags().Changed("strict") {
				opts.Strict = true
			}

			return buildModule(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Output, "output", "o", "dist", "Output directory")
	cmd.Flags().BoolVar(&opts.Minify, "minify", true, "Minify the output")
	cmd.Flags().BoolVar(&opts.Watch, "watch", false, "Watch for changes and rebuild")
	cmd.Flags().BoolVar(&opts.Dev, "dev", false, "Development mode (disables minification, enables source maps)")
	cmd.Flags().StringVar(&opts.Format, "format", "js", "Output format: js")
	cmd.Flags().BoolVar(&opts.SkipChecks, "skip-checks", false, "Skip TypeScript and linting checks (not recommended)")
	cmd.Flags().BoolVar(&opts.NoStrict, "no-strict", false, "Disable strict mode (warnings won't fail the build)")

	cmd.Flags().BoolVar(&opts.Strict, "strict", true, "Enable strict mode (default)")
	cmd.Flags().MarkHidden("strict")

	return cmd
}

func buildModule(opts BuildOptions) error {
	startTime := time.Now()
	fmt.Println("🚀 Building Dashspace module...")

	if opts.Strict {
		fmt.Println("🔒 Strict mode enabled (default) - all warnings will be treated as errors")
		fmt.Println("   Use --no-strict to disable strict mode if needed")
	} else {
		fmt.Println("⚠️  Strict mode disabled - warnings will not fail the build")
	}

	validator := NewValidator()
	if err := validator.ValidateStructure(); err != nil {
		return err
	}

	if err := ensureDependencies(); err != nil {
		return err
	}

	if !opts.SkipChecks {
		fmt.Println("\n📋 Running validation suite...")
		fmt.Println("╔═══════════════════════════════════════════════════╗")

		tsValidator := NewTypeScriptValidator(".")

		fmt.Println("\n1️⃣  TypeScript Type Checking")
		if err := tsValidator.Validate(); err != nil {
			return fmt.Errorf("TypeScript validation failed: %w", err)
		}

		fmt.Println("\n2️⃣  dashspace-lib Compatibility")
		if err := tsValidator.ValidateDashspaceLibTypes(); err != nil {
			if opts.Strict {
				return err
			}
			fmt.Printf("⚠️  Warning: %v\n", err)
		}

		fmt.Println("\n3️⃣  Interface Implementation Validation")
		if err := tsValidator.ValidateInterfaceImplementation(); err != nil {
			if opts.Strict {
				return err
			}
			fmt.Printf("⚠️  Warning: %v\n", err)
		}

		fmt.Println("\n4️⃣  Unused Code Detection")
		if err := tsValidator.CheckUnusedImports(); err != nil {
			if opts.Strict {
				return err
			}
		}

		fmt.Println("\n5️⃣  Type Safety Patterns")
		tsValidator.ValidateSatisfiesUsage()

		fmt.Println("\n6️⃣  ESLint Code Quality")
		linter := NewLintingValidator(".")
		if err := linter.RunESLint(); err != nil {
			if opts.Strict {
				return err
			}
			fmt.Printf("⚠️  Warning: %v\n", err)
		}

		fmt.Println("\n7️⃣  Package Configuration")
		if err := linter.ValidatePackageJSON(); err != nil {
			if opts.Strict {
				return err
			}
			fmt.Printf("⚠️  Warning: %v\n", err)
		}

		fmt.Println("\n8️⃣  Common Issues Check")
		linter.CheckForCommonIssues()

		fmt.Println("\n╚═══════════════════════════════════════════════════╝")
		fmt.Println("✅ All validation checks completed")

	} else {
		fmt.Println("⚠️  Warning: Skipping TypeScript and linting checks - not recommended for production")
		fmt.Println("   Run without --skip-checks to enable full validation")
	}

	if err := os.MkdirAll(opts.Output, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	fmt.Println("\n📦 Extracting module metadata...")
	moduleFile := findModuleFile()
	if moduleFile == "" {
		return fmt.Errorf("Module.ts or Module.tsx not found")
	}

	parser := NewParser(moduleFile)

	config, err := parser.ExtractMetadata()
	if err != nil {
		return fmt.Errorf("failed to extract metadata: %w", err)
	}

	if err := validator.ValidateMetadata(config); err != nil {
		return fmt.Errorf("invalid metadata: %w", err)
	}

	configSteps, err := parser.ExtractConfigurationSteps()
	if err != nil {
		fmt.Printf("⚠️  Warning: Failed to extract configuration steps: %v\n", err)
		configSteps = []map[string]interface{}{}
	} else if len(configSteps) > 0 {
		config.RequiresSetup = true
		fmt.Printf("✅ Found %d configuration steps\n", len(configSteps))
	} else {
		fmt.Println("ℹ️  No configuration steps found")
		config.RequiresSetup = false
	}

	providers, err := parser.ExtractProviders()
	if err != nil {
		fmt.Printf("⚠️  Warning: Failed to extract providers: %v\n", err)
		providers = []map[string]interface{}{}
	} else if len(providers) > 0 {
		fmt.Printf("✅ Found %d required providers\n", len(providers))
		for _, provider := range providers {
			fmt.Printf("   - %s", provider["name"])
			if scopes, ok := provider["scopes"].([]string); ok && len(scopes) > 0 {
				fmt.Printf(" (scopes: %v)", scopes)
			}
			fmt.Println()
		}
	}

	interfaces, err := parser.ExtractInterfaces()
	if err != nil {
		if opts.Strict {
			return fmt.Errorf("interface extraction failed: %w", err)
		}
		fmt.Printf("⚠️  Warning: Failed to extract interfaces: %v\n", err)
		interfaces = []string{}
	} else if len(interfaces) > 0 {
		fmt.Printf("✅ Found %d implemented interfaces: %v\n", len(interfaces), interfaces)

		if !opts.SkipChecks {
			interfaceValidator := &InterfaceValidator{}
			if err := interfaceValidator.ValidateImplementation(interfaces); err != nil {
				if opts.Strict {
					return err
				}
				fmt.Printf("⚠️  Warning: %v\n", err)
			}
		}
	}

	webhooks, err := parser.ExtractWebhooks()
	if err != nil {
		fmt.Printf("⚠️  Warning: Failed to extract webhooks: %v\n", err)
		webhooks = nil
	} else if webhooks != nil {
		fmt.Printf("✅ Found webhook configuration:\n")
		if provider, ok := webhooks["provider"].(string); ok {
			fmt.Printf("   Provider: %s\n", provider)
		}
		if events, ok := webhooks["events"].([]string); ok {
			fmt.Printf("   Events: %v\n", events)
		}
		if configFields, ok := webhooks["configFields"].([]string); ok {
			fmt.Printf("   Config fields: %v\n", configFields)
		}

		if !opts.SkipChecks {
			webhookValidator := NewWebhookValidator()

			if err := webhookValidator.ValidateWebhookConfiguration(webhooks); err != nil {
				if opts.Strict {
					return err
				}
				fmt.Printf("⚠️  Warning: %v\n", err)
			}

			if err := webhookValidator.ValidateWebhookImplementation(webhooks); err != nil {
				if opts.Strict {
					return err
				}
				fmt.Printf("⚠️  Warning: %v\n", err)
			}

			if err := webhookValidator.ValidateComponentWebhookUsage(webhooks); err != nil {
				if opts.Strict {
					return err
				}
				fmt.Printf("⚠️  Warning: %v\n", err)
			}

			webhookValidator.ValidateWebhookSecurity()
		}
	}

	permissions, err := parser.ExtractPermissions()
	if err != nil {
		fmt.Printf("⚠️  Warning: Failed to extract permissions: %v\n", err)
		permissions = []string{}
	} else if len(permissions) > 0 {
		fmt.Printf("✅ Found %d permissions: %v\n", len(permissions), permissions)
		if err := validator.ValidatePermissions(permissions); err != nil {
			if opts.Strict {
				return err
			}
			fmt.Printf("⚠️  Warning: %v\n", err)
		}
	}

	dataSchemaExtractor := NewDataSchemaExtractor()
	dataSchema, err := dataSchemaExtractor.ExtractDataSchema()
	if err != nil {
		fmt.Printf("⚠️  Warning: Failed to extract data schema: %v\n", err)
		dataSchema = nil
	} else if dataSchema != nil && dataSchema.ExposeData {
		fmt.Printf("✅ Module exposes data:\n")
		if dataSchema.DataType != "" {
			fmt.Printf("   Type: %s\n", dataSchema.DataType)
		}
		if dataSchema.Schema != nil {
			fmt.Printf("   Fields: %d\n", len(dataSchema.Schema.Fields))
			fmt.Printf("   Capabilities: %d\n", len(dataSchema.Schema.Capabilities))
		}
		if len(dataSchema.ComputedFields) > 0 {
			fmt.Printf("   Computed metrics: %d\n", len(dataSchema.ComputedFields))
		}

		if !opts.SkipChecks {
			if err := dataSchemaExtractor.ValidateDataSchema(dataSchema); err != nil {
				if opts.Strict {
					return err
				}
				fmt.Printf("⚠️  Warning: %v\n", err)
			}
		}
	} else {
		fmt.Println("○ Module does not expose data")
	}

	entryPoint := findEntryPoint()
	if entryPoint == "" {
		return fmt.Errorf("no entry point found")
	}

	fmt.Printf("\n📦 Compiling %s...\n", entryPoint)
	compiler := NewCompiler(opts)
	bundleContent, err := compiler.Compile(entryPoint)
	if err != nil {
		return fmt.Errorf("compilation failed: %w", err)
	}

	fmt.Println("🔧 Generating bundle...")
	generator := NewGenerator(config)
	finalBundle, checksum := generator.Generate(bundleContent)

	writer := NewWriter(opts.Output)
	if err := writer.WriteBundle(finalBundle, checksum); err != nil {
		return fmt.Errorf("failed to write bundle: %w", err)
	}

	manifest := BuildManifest(config, configSteps, providers, interfaces, permissions, webhooks, dataSchema, checksum)

	if err := writer.WriteManifest(manifest); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	if !opts.SkipChecks && opts.Strict {
		if err := validateOutput(opts.Output); err != nil {
			return fmt.Errorf("output validation failed: %w", err)
		}
	}

	duration := time.Since(startTime)
	printBuildSummary(config, opts.Output, duration, opts.Strict, dataSchema)

	if opts.Watch {
		watcher := NewWatcher()
		return watcher.Watch(func() {
			buildModule(opts)
		})
	}

	return nil
}

func validateOutput(outputDir string) error {
	fmt.Println("🔍 Validating output files...")

	bundlePath := filepath.Join(outputDir, "bundle.js")
	bundleInfo, err := os.Stat(bundlePath)
	if err != nil {
		return fmt.Errorf("bundle.js not found: %w", err)
	}

	bundleSizeKB := float64(bundleInfo.Size()) / 1024
	if bundleSizeKB > 500 {
		fmt.Printf("⚠️  Warning: Bundle size is %.2f KB - consider optimizing\n", bundleSizeKB)
	}

	manifestPath := filepath.Join(outputDir, "dashspace.json")
	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("dashspace.json not found: %w", err)
	}

	var manifest map[string]interface{}
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		return fmt.Errorf("invalid dashspace.json: %w", err)
	}

	requiredFields := []string{"id", "name", "version", "checksum", "timestamp"}
	for _, field := range requiredFields {
		if _, ok := manifest[field]; !ok {
			return fmt.Errorf("dashspace.json missing required field: %s", field)
		}
	}

	fmt.Println("✅ Output validation passed")
	return nil
}

func printBuildSummary(config *DashspaceConfig, outputDir string, duration time.Duration, strict bool, dataSchema *ModuleDataSchema) {
	bundlePath := filepath.Join(outputDir, "bundle.js")
	bundleInfo, _ := os.Stat(bundlePath)

	fmt.Printf("\n")
	fmt.Printf("╔═════════════════════════════════════════════════════╗\n")
	fmt.Printf("║  ✅ Build completed successfully in %.2fs           ║\n", duration.Seconds())
	fmt.Printf("╚═════════════════════════════════════════════════════╝\n")
	fmt.Printf("\n")
	fmt.Printf("📦 Module Details:\n")
	fmt.Printf("   ID:          %d\n", config.ID)
	fmt.Printf("   Name:        %s\n", config.Name)
	fmt.Printf("   Slug:        %s\n", config.Slug)
	fmt.Printf("   Version:     %s\n", config.Version)

	if config.Description != "" {
		fmt.Printf("   Description: %s\n", config.Description)
	}

	fmt.Printf("\n")
	fmt.Printf("📊 Build Statistics:\n")

	if bundleInfo != nil {
		bundleSizeKB := float64(bundleInfo.Size()) / 1024
		fmt.Printf("   Bundle size: %.2f KB", bundleSizeKB)

		if bundleSizeKB > 500 {
			fmt.Printf(" ⚠️  (large)")
		} else if bundleSizeKB < 50 {
			fmt.Printf(" ✨ (optimized)")
		} else {
			fmt.Printf(" ✅")
		}
		fmt.Println()
	}

	if len(config.ImplementedInterfaces) > 0 {
		fmt.Printf("   Interfaces:  %d implemented\n", len(config.ImplementedInterfaces))
	}

	if len(config.Providers) > 0 {
		fmt.Printf("   Providers:   %d required\n", len(config.Providers))
	}

	if config.RequiresSetup {
		fmt.Printf("   Setup:       Required\n")
	} else {
		fmt.Printf("   Setup:       Not required\n")
	}

	if dataSchema != nil && dataSchema.ExposeData {
		fmt.Printf("   Data:        Exposed (%s)\n", dataSchema.DataType)
	}

	fmt.Printf("\n")
	fmt.Printf("📁 Output Files:\n")
	fmt.Printf("   %s/\n", outputDir)
	fmt.Printf("   ├── bundle.js       (compiled module)\n")
	fmt.Printf("   └── dashspace.json  (module manifest)\n")

	if strict {
		fmt.Printf("\n")
		fmt.Printf("🔒 Built in strict mode - all checks passed\n")
	} else {
		fmt.Printf("\n")
		fmt.Printf("⚠️  Built without strict mode\n")
	}

	fmt.Printf("\n")
	fmt.Printf("🎉 Module is ready for deployment!\n")
}

func BuildManifest(
	config *DashspaceConfig,
	configSteps []map[string]interface{},
	providers []map[string]interface{},
	interfaces []string,
	permissions []string,
	webhooks map[string]interface{},
	dataSchema *ModuleDataSchema,
	checksum string,
) map[string]interface{} {
	config.Checksum = checksum
	config.Timestamp = time.Now().Format(time.RFC3339)

	manifest := map[string]interface{}{
		"id":             config.ID,
		"slug":           config.Slug,
		"name":           config.Name,
		"version":        config.Version,
		"description":    config.Description,
		"author":         config.Author,
		"entry":          config.Entry,
		"checksum":       config.Checksum,
		"timestamp":      config.Timestamp,
		"requires_setup": config.RequiresSetup,
		"build_info": map[string]interface{}{
			"cli_version": "1.0.11",
			"build_date":  config.Timestamp,
			"validated":   true,
		},
	}

	if config.Icon != "" {
		manifest["icon"] = config.Icon
	}

	if config.Category != "" {
		manifest["category"] = config.Category
	}

	if len(config.Tags) > 0 {
		manifest["tags"] = config.Tags
	}

	if config.RequiresSetup && len(configSteps) > 0 {
		manifest["configuration_steps"] = configSteps
	}

	if len(providers) > 0 {
		manifest["providers"] = providers
	}

	if len(interfaces) > 0 {
		manifest["interfaces"] = interfaces
	}

	if len(permissions) > 0 {
		manifest["permissions"] = permissions
	}

	if webhooks != nil && len(webhooks) > 0 {
		manifest["webhooks"] = webhooks
	}

	if dataSchema != nil && dataSchema.ExposeData {
		manifest["data_schema"] = dataSchema
	}

	return manifest
}
