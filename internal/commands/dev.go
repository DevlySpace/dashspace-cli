package commands

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
)

func NewDevCmd() *cobra.Command {
	var (
		port   string
		noOpen bool
	)

	cmd := &cobra.Command{
		Use:   "dev",
		Short: "Start development server with hot-reload",
		Long:  "Start a local development server to test your module with live reloading",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDevServer(port, noOpen)
		},
	}

	cmd.Flags().StringVarP(&port, "port", "p", "3000", "Port to run the dev server")
	cmd.Flags().BoolVar(&noOpen, "no-open", false, "Don't open browser automatically")

	return cmd
}

func runDevServer(port string, noOpen bool) error {
	fmt.Println("🚀 Starting Dashspace development server...")

	// Check if esbuild is installed, if not install it
	if _, err := exec.LookPath("esbuild"); err != nil {
		fmt.Println("📦 Installing esbuild...")
		installCmd := exec.Command("npm", "install", "-D", "esbuild")
		installCmd.Stdout = os.Stdout
		installCmd.Stderr = os.Stderr
		if err := installCmd.Run(); err != nil {
			return fmt.Errorf("failed to install esbuild: %v", err)
		}
	}

	// Initial build
	if err := buildForDev(); err != nil {
		return fmt.Errorf("initial build failed: %v", err)
	}

	// Start file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %v", err)
	}
	defer watcher.Close()

	// Watch source files
	watchFiles := []string{
		"Module.ts",
		"Module.tsx",
		"Component.tsx",
		"types.ts",
	}

	for _, file := range watchFiles {
		if _, err := os.Stat(file); err == nil {
			if err := watcher.Add(file); err != nil {
				fmt.Printf("⚠️  Failed to watch %s: %v\n", file, err)
			}
		}
	}

	// Start rebuild goroutine
	go func() {
		var lastBuild time.Time
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					if time.Since(lastBuild) > 100*time.Millisecond {
						fmt.Printf("📝 File changed: %s\n", event.Name)
						fmt.Println("🔄 Rebuilding...")
						if err := buildForDev(); err != nil {
							fmt.Printf("❌ Build failed: %v\n", err)
						} else {
							fmt.Println("✅ Build successful")
						}
						lastBuild = time.Now()
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Printf("❌ Watcher error: %v\n", err)
			}
		}
	}()

	// Setup HTTP server
	mux := http.NewServeMux()

	// Serve the built bundle
	mux.HandleFunc("/bundle.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		http.ServeFile(w, r, ".dev/bundle.js")
	})

	// Serve main HTML page
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		serveDevHTML(w, r)
	})

	serverURL := fmt.Sprintf("http://localhost:%s", port)
	fmt.Printf("\n✨ Development server running at %s\n", serverURL)
	fmt.Println("👀 Watching for file changes...")
	fmt.Println("\nPress Ctrl+C to stop")

	// Open browser
	if !noOpen {
		go func() {
			time.Sleep(1 * time.Second)
			openBrowser(serverURL)
		}()
	}

	// Start server
	return http.ListenAndServe(":"+port, mux)
}

func buildForDev() error {
	// Create .dev directory if it doesn't exist
	if err := os.MkdirAll(".dev", 0755); err != nil {
		return err
	}

	// Generate index.ts if needed
	indexGenerated := false
	if _, err := os.Stat("index.ts"); err != nil {
		if err := generateIndexFileForDev(); err != nil {
			return fmt.Errorf("failed to generate index: %v", err)
		}
		indexGenerated = true
		defer func() {
			if indexGenerated {
				os.Remove("index.ts")
			}
		}()
	}

	// Bundle with esbuild (handles TypeScript, JSX, and ES modules automatically)
	cmd := exec.Command("npx", "esbuild",
		"index.ts",
		"--bundle",
		"--format=esm",
		"--platform=browser",
		"--target=es2020",
		"--jsx=automatic",
		"--loader:.ts=ts",
		"--loader:.tsx=tsx",
		"--outfile=.dev/bundle.js",
		"--sourcemap",
		"--external:react",
		"--external:react-dom",
		"--external:react-dom/client",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("esbuild failed: %s", string(output))
	}

	return nil
}

func generateIndexFileForDev() error {
	moduleFile := ""
	if _, err := os.Stat("Module.ts"); err == nil {
		moduleFile = "Module"
	} else if _, err := os.Stat("Module.tsx"); err == nil {
		moduleFile = "Module"
	} else {
		return fmt.Errorf("Module.ts or Module.tsx not found")
	}

	hasComponent := false
	componentFile := ""
	if _, err := os.Stat("Component.tsx"); err == nil {
		hasComponent = true
		componentFile = "Component"
	}

	// Generate a complete index that mounts the React component
	indexContent := ""

	if hasComponent {
		indexContent = fmt.Sprintf(`import ModuleClass from './%s';
import Component from './%s';
import React from 'react';
import { createRoot } from 'react-dom/client';

// Export module for build system
export default ModuleClass;

// Dev mode: Auto-initialize and mount
if (typeof window !== 'undefined') {
    // Wait for DOM to be ready
    const init = () => {
        console.log('Initializing module and mounting component...');
        
        // Initialize module
        if (window.DashspaceLib) {
            window.DashspaceLib.autoInitialize(ModuleClass);
        }
        
        // Mount React component
        const rootElement = document.getElementById('root');
        if (rootElement) {
            console.log('Mounting React component to #root');
            try {
                const root = createRoot(rootElement);
                root.render(React.createElement(Component));
                console.log('Component mounted successfully');
            } catch (err) {
                console.error('Failed to mount component:', err);
                rootElement.innerHTML = '<div style="padding: 2rem; color: red;">Failed to mount component: ' + err.message + '</div>';
            }
        } else {
            console.error('Root element not found');
        }
    };
    
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        // DOM is already loaded
        setTimeout(init, 100); // Small delay to ensure everything is ready
    }
}
`, moduleFile, componentFile)
	} else {
		indexContent = fmt.Sprintf(`import ModuleClass from './%s';
export default ModuleClass;

if (typeof window !== 'undefined' && window.DashspaceLib) {
    window.DashspaceLib.autoInitialize(ModuleClass);
}
`, moduleFile)
	}

	return os.WriteFile("index.ts", []byte(indexContent), 0644)
}

func serveDevHTML(w http.ResponseWriter, r *http.Request) {
	htmlTemplate := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Dashspace Module Dev</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <style>
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: #f3f4f6;
        }
        .dev-banner {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 1rem;
            text-align: center;
            font-weight: 500;
            position: relative;
        }
        .config-btn {
            position: absolute;
            right: 1rem;
            top: 50%;
            transform: translateY(-50%);
            background: rgba(255,255,255,0.2);
            padding: 0.5rem 1rem;
            border-radius: 0.25rem;
            border: none;
            cursor: pointer;
            color: white;
            font-weight: 500;
        }
        .config-btn:hover {
            background: rgba(255,255,255,0.3);
        }
        .config-panel {
            background: white;
            padding: 1.5rem;
            border-bottom: 1px solid #e5e7eb;
            display: none;
        }
        .config-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 1rem;
            max-width: 1200px;
            margin: 0 auto;
        }
        .config-field {
            display: flex;
            flex-direction: column;
        }
        .config-field label {
            font-size: 0.875rem;
            margin-bottom: 0.25rem;
            font-weight: 500;
        }
        .config-field input, .config-field select {
            padding: 0.5rem;
            border: 1px solid #d1d5db;
            border-radius: 0.25rem;
            font-size: 0.875rem;
        }
        .apply-btn {
            margin-top: 1rem;
            background: #6366f1;
            color: white;
            padding: 0.5rem 1.5rem;
            border-radius: 0.25rem;
            border: none;
            cursor: pointer;
            font-weight: 500;
        }
        .apply-btn:hover {
            background: #5558e3;
        }
        .module-container {
            max-width: 1200px;
            margin: 2rem auto;
            background: white;
            border-radius: 0.5rem;
            box-shadow: 0 1px 3px rgba(0,0,0,0.1);
            overflow: hidden;
            min-height: 400px;
        }
    </style>
</head>
<body>
    <div class="dev-banner">
        🚀 Dashspace Development Mode
        <button class="config-btn" onclick="toggleConfig()">
            ⚙️ Settings
        </button>
    </div>
    
    <div id="config-panel" class="config-panel">
        <div style="max-width: 1200px; margin: 0 auto;">
            <h3 style="margin: 0 0 1rem 0; font-weight: 600;">Module Configuration</h3>
            <div id="config-fields" class="config-grid">
                <!-- Fields will be dynamically generated here -->
            </div>
            <button class="apply-btn" onclick="applyConfig()">
                Apply Configuration
            </button>
        </div>
    </div>
    
    <div class="module-container">
        <div id="root"></div>
    </div>
    
    <script type="importmap">
    {
        "imports": {
            "react": "https://esm.sh/react@18",
            "react-dom": "https://esm.sh/react-dom@18",
            "react-dom/client": "https://esm.sh/react-dom@18/client"
        }
    }
    </script>
    
    <script>
        // Configuration management
        let currentConfig = {};
        let moduleFields = [];
        
        function toggleConfig() {
            const panel = document.getElementById('config-panel');
            panel.style.display = panel.style.display === 'none' ? 'block' : 'none';
        }
        
        function renderConfigFields(fields) {
            const container = document.getElementById('config-fields');
            container.innerHTML = '';
            
            fields.forEach(field => {
                const div = document.createElement('div');
                div.className = 'config-field';
                
                const label = document.createElement('label');
                label.textContent = field.label || field.name;
                div.appendChild(label);
                
                let input;
                if (field.type === 'select' || field.type === 'multiselect') {
                    input = document.createElement('select');
                    if (field.options) {
                        field.options.forEach(opt => {
                            const option = document.createElement('option');
                            option.value = opt.value;
                            option.textContent = opt.label;
                            input.appendChild(option);
                        });
                    }
                } else if (field.type === 'boolean') {
                    input = document.createElement('input');
                    input.type = 'checkbox';
                } else if (field.type === 'number') {
                    input = document.createElement('input');
                    input.type = 'number';
                    if (field.min !== undefined) input.min = field.min;
                    if (field.max !== undefined) input.max = field.max;
                    if (field.step !== undefined) input.step = field.step;
                } else {
                    input = document.createElement('input');
                    input.type = 'text';
                }
                
                input.id = 'config-' + field.name;
                input.value = currentConfig[field.name] || field.defaultValue || '';
                if (field.type === 'boolean') {
                    input.checked = currentConfig[field.name] || field.defaultValue || false;
                }
                
                div.appendChild(input);
                container.appendChild(div);
            });
        }
        
        function applyConfig() {
            // Collect values from form
            const newConfig = {};
            moduleFields.forEach(field => {
                const input = document.getElementById('config-' + field.name);
                if (input) {
                    if (field.type === 'boolean') {
                        newConfig[field.name] = input.checked;
                    } else if (field.type === 'number') {
                        newConfig[field.name] = parseInt(input.value);
                    } else {
                        newConfig[field.name] = input.value;
                    }
                }
            });
            
            console.log('Applying config:', newConfig);
            
            // Use the new DashspaceLib method to apply configuration
            if (window.DashspaceLib) {
                try {
                    const success = window.DashspaceLib.applyConfiguration(newConfig);
                    
                    if (success) {
                        console.log('Configuration applied successfully');
                        
                        // Hide the config panel
                        toggleConfig();
                        
                        // Trigger a re-render by dispatching a custom event
                        window.dispatchEvent(new CustomEvent('configUpdated', { 
                            detail: newConfig 
                        }));
                        
                    } else {
                        throw new Error('Failed to apply configuration');
                    }
                    
                } catch (error) {
                    console.error('Failed to apply configuration:', error);
                    alert('Erreur lors de l\'application de la configuration: ' + error.message);
                }
            } else {
                console.warn('DashspaceLib not available, will reload page');
                location.reload();
            }
        }
        
        // Load saved config
        const savedConfig = sessionStorage.getItem('devConfig');
        if (savedConfig) {
            currentConfig = JSON.parse(savedConfig);
        }
        
        // Mock Dashspace environment
        window.modly = {
            config: {
                get: function(slug) {
                    console.log('Config requested for slug:', slug, 'returning:', currentConfig);
                    return currentConfig;
                }
            },
            providers: {
                call: async function(provider, request) {
                    console.log('Provider call:', provider, request);
                    
                    // Proxy GitHub API for development
                    if (provider === 'github' && request.endpoint) {
                        try {
                            const url = 'https://api.github.com' + request.endpoint;
                            const response = await fetch(url, {
                                headers: {
                                    'Accept': 'application/vnd.github.v3+json'
                                }
                            });
                            const data = await response.json();
                            return { data };
                        } catch (err) {
                            console.error('API call failed:', err);
                            return { data: [] };
                        }
                    }
                    
                    return { data: [] };
                },
                isAuthenticated: function() { return false; },
                list: function() { return []; }
            }
        };

        // Create global DashspaceLib before loading the bundle
        window.DashspaceLib = {
            instance: null,
            initialized: false,
            configApplied: false,
            
            setModule(module) {
                this.instance = module;
                
                // Extract fields from configuration steps for the UI
                if (!moduleFields.length) {
                    const steps = module.getConfigurationSteps();
                    steps.forEach(step => {
                        if (step.fields) {
                            moduleFields.push(...step.fields);
                        }
                    });
                    
                    // Set default values if no config saved
                    if (Object.keys(currentConfig).length === 0) {
                        moduleFields.forEach(field => {
                            if (field.defaultValue !== undefined) {
                                currentConfig[field.name] = field.defaultValue;
                            }
                        });
                        
                        // Apply default config to module immediately
                        if (Object.keys(currentConfig).length > 0) {
                            console.log('Setting default config to module:', currentConfig);
                            module.setConfig(currentConfig);
                            this.configApplied = true;
                        }
                    }
                    
                    // Render config fields in UI
                    renderConfigFields(moduleFields);
                }
            },
            
            isSetupCompleted() {
                // Le module doit exister ET avoir une configuration valide
                const hasInstance = !!this.instance;
                const moduleSetup = this.instance && this.instance.isSetupCompleted();
                const hasValidConfig = Object.keys(currentConfig).length > 0;
                
                const result = hasInstance && moduleSetup && hasValidConfig;
                
                console.log('DashspaceLib.isSetupCompleted check:', {
                    hasInstance,
                    moduleSetup,
                    hasValidConfig,
                    currentConfigKeys: Object.keys(currentConfig),
                    result
                });
                
                return result;
            },
            
            getConfig() {
                // Toujours retourner currentConfig qui est la source de vérité
                console.log('DashspaceLib.getConfig returning currentConfig:', currentConfig);
                return { ...currentConfig };
            },
            
            async callProvider(provider, options) {
                if (window.modly?.providers) {
                    const result = await window.modly.providers.call(provider, options);
                    return result.data || result;
                }
                throw new Error('Provider not available');
            },
            
            // Nouvelle méthode pour appliquer la configuration
            applyConfiguration(newConfig) {
                console.log('DashspaceLib.applyConfiguration called with:', newConfig);
                
                // Mettre à jour currentConfig
                Object.assign(currentConfig, newConfig);
                
                // Sauvegarder dans sessionStorage
                sessionStorage.setItem('devConfig', JSON.stringify(currentConfig));
                
                // Appliquer au module si il existe
                if (this.instance) {
                    try {
                        this.instance.setConfig(currentConfig);
                        this.configApplied = true;
                        console.log('Configuration successfully applied to module');
                        return true;
                    } catch (error) {
                        console.error('Failed to apply config to module:', error);
                        return false;
                    }
                }
                
                return true;
            },
            
            async initializeModule(ModuleClass, config) {
                console.log('DashspaceLib: initializeModule called');
                
                if (this.initialized) {
                    console.log('DashspaceLib: Already initialized, returning existing instance');
                    return this.instance;
                }
                
                const module = new ModuleClass();
                
                // Get default config from steps
                const defaultConfig = {};
                const steps = module.getConfigurationSteps();
                for (const step of steps) {
                    if (step.fields) {
                        for (const field of step.fields) {
                            if (field.defaultValue !== undefined) {
                                defaultConfig[field.name] = field.defaultValue;
                            }
                        }
                    }
                }
                
                // Merge avec la configuration sauvegardée ou fournie
                if (Object.keys(currentConfig).length === 0) {
                    Object.assign(currentConfig, defaultConfig);
                }
                
                if (config) {
                    Object.assign(currentConfig, config);
                }
                
                console.log('DashspaceLib: Final config for module:', currentConfig);
                
                // Appliquer la configuration au module
                module.setConfig(currentConfig);
                
                if (module.start) {
                    await module.start();
                } else if (module.initialize) {
                    await module.initialize();
                }
                
                this.setModule(module);
                this.initialized = true;
                this.configApplied = true;
                
                console.log('DashspaceLib: Module initialized successfully');
                return module;
            },
            
            async autoInitialize(ModuleClass) {
                console.log('DashspaceLib: autoInitialize called');
                if (typeof window !== 'undefined' && !this.initialized) {
                    try {
                        await this.initializeModule(ModuleClass);
                        console.log('DashspaceLib: Auto-initialization completed');
                    } catch (error) {
                        console.error('DashspaceLib: Failed to auto-initialize module:', error);
                        throw error;
                    }
                }
            }
        };
    </script>
    
    <script type="module">
        // Show loading state
        document.getElementById('root').innerHTML = '<div style="padding: 2rem; text-align: center; color: #666;">Loading module...</div>';
        
        // Load the bundled module
        import('/bundle.js').then(() => {
            console.log('Bundle loaded successfully');
        }).catch(err => {
            console.error('Failed to load bundle:', err);
            document.getElementById('root').innerHTML = 
                '<div style="padding: 2rem; color: red;">Failed to load module: ' + err.message + '</div>';
        });
    </script>
</body>
</html>`

	tmpl, err := template.New("dev").Parse(htmlTemplate)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	tmpl.Execute(w, nil)
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	}
	if cmd != nil {
		cmd.Start()
	}
}
