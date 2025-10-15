package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func NewCreateCmd() *cobra.Command {
	var (
		templateType string
		typescript   bool
	)

	cmd := &cobra.Command{
		Use:   "create [module-name]",
		Short: "Create a new Dashspace module",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			moduleName := args[0]
			return createModule(moduleName, templateType, typescript)
		},
	}

	cmd.Flags().StringVarP(&templateType, "template", "t", "react", "Template type (react, vanilla, chart, list)")
	cmd.Flags().BoolVar(&typescript, "typescript", true, "Use TypeScript with dashspace-lib")

	return cmd
}

func createModule(name, templateType string, useTypescript bool) error {
	fmt.Printf("🚀 Creating Dashspace module '%s'\n", name)

	// Check if directory exists
	if _, err := os.Stat(name); err == nil {
		return fmt.Errorf("directory '%s' already exists", name)
	}

	// Create directory
	if err := os.MkdirAll(name, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	// Create src directory for TypeScript
	srcDir := filepath.Join(name, "src")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		return fmt.Errorf("failed to create src directory: %v", err)
	}

	if useTypescript {
		// Generate TypeScript module with dashspace-lib
		if err := generateTypescriptModule(name, templateType); err != nil {
			return err
		}
	} else {
		// Generate legacy JavaScript module
		if err := generateJavaScriptModule(name, templateType); err != nil {
			return err
		}
	}

	fmt.Printf("\n✅ Module '%s' created successfully!\n", name)
	fmt.Println("\n📝 Next steps:")
	fmt.Printf("   cd %s\n", name)
	fmt.Println("   npm install")
	fmt.Println("   npm run build")
	fmt.Println("   dashspace publish")

	return nil
}

func generateTypescriptModule(name, templateType string) error {
	// Generate package.json
	packageJSON := generatePackageJSON(name)
	if err := writeFile(filepath.Join(name, "package.json"), packageJSON); err != nil {
		return err
	}

	// Generate tsconfig.json
	tsConfig := generateTSConfig()
	if err := writeFile(filepath.Join(name, "tsconfig.json"), tsConfig); err != nil {
		return err
	}

	// Generate module source
	moduleSource := generateModuleSource(name, templateType)
	if err := writeFile(filepath.Join(name, "src", "index.tsx"), moduleSource); err != nil {
		return err
	}

	// Generate .gitignore
	gitignore := generateGitignore()
	if err := writeFile(filepath.Join(name, ".gitignore"), gitignore); err != nil {
		return err
	}

	// Generate README
	readme := generateReadme(name, templateType)
	if err := writeFile(filepath.Join(name, "README.md"), readme); err != nil {
		return err
	}

	return nil
}

func generateJavaScriptModule(name, templateType string) error {
	// Legacy support - generate simple JS module
	// This is the old devly.json based approach
	return fmt.Errorf("JavaScript modules are deprecated. Use TypeScript with --typescript flag")
}

func generatePackageJSON(name string) string {
	return fmt.Sprintf(`{
  "name": "%s",
  "version": "1.0.0",
  "description": "Dashspace module",
  "main": "dist/index.js",
  "scripts": {
    "build": "tsc",
    "dev": "tsc --watch",
    "test": "echo \"No tests yet\""
  },
  "dependencies": {
    "dashspace-lib": "^1.0.0",
    "react": "^18.3.1",
    "react-dom": "^18.3.1"
  },
  "devDependencies": {
    "@types/node": "^20.0.0",
    "@types/react": "^18.3.0",
    "@types/react-dom": "^18.3.0",
    "typescript": "^5.0.0"
  }
}`, name)
}

func generateTSConfig() string {
	return `{
  "compilerOptions": {
    "target": "ES2020",
    "module": "commonjs",
    "lib": ["ES2020", "DOM"],
    "jsx": "react",
    "outDir": "./dist",
    "rootDir": "./src",
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true,
    "declaration": true,
    "declarationMap": true,
    "moduleResolution": "node",
    "resolveJsonModule": true
  },
  "include": ["src/**/*"],
  "exclude": ["node_modules", "dist"]
}`
}

func generateModuleSource(name, templateType string) string {
	className := toPascalCase(name) + "Module"

	switch templateType {
	case "chart":
		return generateChartModule(name, className)
	case "list":
		return generateListModule(name, className)
	default:
		return generateBasicModule(name, className)
	}
}

func generateBasicModule(name, className string) string {
	return fmt.Sprintf(`import React, { useState, useEffect, useCallback } from 'react';
import {
    DashspaceModule,
    ConfigurationStep,
    TextField,
    NumberField,
    SelectField,
    BooleanField
} from 'dashspace-lib';

interface ModuleData {
    // Define your data structure here
    items: any[];
}

export class %s extends DashspaceModule<ModuleData> {
    constructor() {
        super({
            id: '%s',
            name: '%s',
            version: '1.0.0',
            description: 'A Dashspace module',
            author: 'Your Name',
            icon: '📦',
            category: 'General'
        });
    }

    defineConfigurationSteps(): ConfigurationStep[] {
        return [
            new ConfigurationStep({
                id: 'general',
                title: 'General Configuration',
                order: 1,
                fields: [
                    new TextField({
                        name: 'title',
                        label: 'Module Title',
                        defaultValue: '%s',
                        validation: { required: true }
                    }),
                    new NumberField({
                        name: 'refreshRate',
                        label: 'Refresh Rate (seconds)',
                        defaultValue: 60,
                        min: 10,
                        max: 3600
                    })
                ]
            })
        ];
    }

    async initialize(): Promise<void> {
        console.log('%s module initialized');
    }

    async callProvider(endpoint: string, params?: Record<string, any>): Promise<ModuleData> {
        // Implement your API calls here
        return { items: [] };
    }

    async cleanup(): Promise<void> {
        console.log('%s module cleaned up');
    }
}

const %sComponent: React.FC = () => {
    const [data, setData] = useState<any[]>([]);
    const [loading, setLoading] = useState(false);

    const setupComplete = window.DASHSPACE_LIB?.isSetupCompleted() || false;
    const config = window.DASHSPACE_LIB?.getConfig() || {};

    const loadData = useCallback(async () => {
        if (!setupComplete || !window.DASHSPACE_LIB) return;

        setLoading(true);
        try {
            const response = await window.DASHSPACE_LIB.callProvider('/data');
            setData(response.items || []);
        } catch (error) {
            console.error('Error loading data:', error);
        } finally {
            setLoading(false);
        }
    }, [setupComplete]);

    useEffect(() => {
        if (setupComplete) {
            loadData();
        }
    }, [setupComplete, loadData]);

    if (!setupComplete) {
        return <div className="p-4 text-center">Configuring...</div>;
    }

    return (
        <div className="p-4">
            <h2 className="text-xl font-bold mb-4">{config.title || '%s'}</h2>
            {loading && <div>Loading...</div>}
            <div>
                {data.map((item, index) => (
                    <div key={index}>{JSON.stringify(item)}</div>
                ))}
            </div>
        </div>
    );
};

export default %sComponent;`,
		className, name, strings.Title(strings.ReplaceAll(name, "-", " ")),
		strings.Title(strings.ReplaceAll(name, "-", " ")),
		className, className,
		className, strings.Title(strings.ReplaceAll(name, "-", " ")),
		className)
}

func generateChartModule(name, className string) string {
	return fmt.Sprintf(`import React, { useState, useEffect, useCallback } from 'react';
import {
    DashspaceModule,
    ConfigurationStep,
    TextField,
    NumberField,
    SelectField
} from 'dashspace-lib';

interface ChartData {
    labels: string[];
    values: number[];
}

export class %s extends DashspaceModule<ChartData> {
    constructor() {
        super({
            id: '%s',
            name: '%s',
            version: '1.0.0',
            description: 'Chart module for data visualization',
            author: 'Your Name',
            icon: '📊',
            category: 'Visualization'
        });
    }

    defineConfigurationSteps(): ConfigurationStep[] {
        return [
            new ConfigurationStep({
                id: 'chart-config',
                title: 'Chart Configuration',
                order: 1,
                fields: [
                    new TextField({
                        name: 'title',
                        label: 'Chart Title',
                        defaultValue: 'My Chart',
                        validation: { required: true }
                    }),
                    new SelectField({
                        name: 'chartType',
                        label: 'Chart Type',
                        defaultValue: 'bar',
                        options: [
                            { value: 'bar', label: 'Bar Chart' },
                            { value: 'line', label: 'Line Chart' },
                            { value: 'pie', label: 'Pie Chart' }
                        ]
                    }),
                    new NumberField({
                        name: 'maxItems',
                        label: 'Maximum Items',
                        defaultValue: 10,
                        min: 5,
                        max: 50
                    })
                ]
            })
        ];
    }

    async initialize(): Promise<void> {
        console.log('Chart module initialized');
    }

    async callProvider(endpoint: string, params?: Record<string, any>): Promise<ChartData> {
        // Mock data - replace with actual API call
        return {
            labels: ['Jan', 'Feb', 'Mar', 'Apr', 'May'],
            values: [400, 300, 600, 800, 500]
        };
    }

    async cleanup(): Promise<void> {
        console.log('Chart module cleaned up');
    }
}

const ChartComponent: React.FC = () => {
    const [chartData, setChartData] = useState<ChartData | null>(null);
    const [loading, setLoading] = useState(false);

    const setupComplete = window.DASHSPACE_LIB?.isSetupCompleted() || false;
    const config = window.DASHSPACE_LIB?.getConfig() || {};

    const loadData = useCallback(async () => {
        if (!setupComplete || !window.DASHSPACE_LIB) return;

        setLoading(true);
        try {
            const data = await window.DASHSPACE_LIB.callProvider('/chart-data');
            setChartData(data);
        } catch (error) {
            console.error('Error loading chart data:', error);
        } finally {
            setLoading(false);
        }
    }, [setupComplete]);

    useEffect(() => {
        if (setupComplete) {
            loadData();
        }
    }, [setupComplete, loadData]);

    if (!setupComplete) {
        return <div className="p-4 text-center">Configuring...</div>;
    }

    const maxValue = chartData ? Math.max(...chartData.values) : 100;

    return (
        <div className="p-4 bg-white rounded-lg shadow">
            <h2 className="text-xl font-bold mb-4">{config.title || 'Chart'}</h2>

            {loading && <div>Loading chart...</div>}

            {chartData && (
                <div className="flex items-end space-x-2 h-32">
                    {chartData.values.map((value, index) => (
                        <div key={index} className="flex-1 flex flex-col items-center">
                            <div
                                className="bg-blue-500 w-full rounded-t"
                                style={{ height: `+"`${(value / maxValue) * 100}%`"+` }}
                            ></div>
                            <span className="text-xs mt-1">{chartData.labels[index]}</span>
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
};

export default ChartComponent;`, className, name, strings.Title(strings.ReplaceAll(name, "-", " ")))
}

func generateListModule(name, className string) string {
	return fmt.Sprintf(`import React, { useState, useEffect, useCallback } from 'react';
import {
    DashspaceModule,
    ConfigurationStep,
    TextField,
    NumberField,
    BooleanField
} from 'dashspace-lib';

interface ListItem {
    id: string;
    title: string;
    description: string;
    status: 'active' | 'inactive' | 'pending';
}

export class %s extends DashspaceModule<ListItem[]> {
    constructor() {
        super({
            id: '%s',
            name: '%s',
            version: '1.0.0',
            description: 'List module for displaying items',
            author: 'Your Name',
            icon: '📋',
            category: 'Display'
        });
    }

    defineConfigurationSteps(): ConfigurationStep[] {
        return [
            new ConfigurationStep({
                id: 'list-config',
                title: 'List Configuration',
                order: 1,
                fields: [
                    new TextField({
                        name: 'title',
                        label: 'List Title',
                        defaultValue: 'My List',
                        validation: { required: true }
                    }),
                    new NumberField({
                        name: 'itemsPerPage',
                        label: 'Items per page',
                        defaultValue: 10,
                        min: 5,
                        max: 50
                    }),
                    new BooleanField({
                        name: 'showSearch',
                        label: 'Show search bar',
                        defaultValue: true
                    })
                ]
            })
        ];
    }

    async initialize(): Promise<void> {
        console.log('List module initialized');
    }

    async callProvider(endpoint: string, params?: Record<string, any>): Promise<ListItem[]> {
        // Mock data - replace with actual API call
        return [
            { id: '1', title: 'Item 1', description: 'Description 1', status: 'active' },
            { id: '2', title: 'Item 2', description: 'Description 2', status: 'pending' },
            { id: '3', title: 'Item 3', description: 'Description 3', status: 'inactive' }
        ];
    }

    async cleanup(): Promise<void> {
        console.log('List module cleaned up');
    }
}

const ListComponent: React.FC = () => {
    const [items, setItems] = useState<ListItem[]>([]);
    const [searchQuery, setSearchQuery] = useState('');
    const [loading, setLoading] = useState(false);

    const setupComplete = window.DASHSPACE_LIB?.isSetupCompleted() || false;
    const config = window.DASHSPACE_LIB?.getConfig() || {};

    const loadItems = useCallback(async () => {
        if (!setupComplete || !window.DASHSPACE_LIB) return;

        setLoading(true);
        try {
            const data = await window.DASHSPACE_LIB.callProvider('/items');
            setItems(data);
        } catch (error) {
            console.error('Error loading items:', error);
        } finally {
            setLoading(false);
        }
    }, [setupComplete]);

    useEffect(() => {
        if (setupComplete) {
            loadItems();
        }
    }, [setupComplete, loadItems]);

    const filteredItems = items.filter(item =>
        item.title.toLowerCase().includes(searchQuery.toLowerCase())
    );

    if (!setupComplete) {
        return <div className="p-4 text-center">Configuring...</div>;
    }

    return (
        <div className="p-4">
            <h2 className="text-xl font-bold mb-4">{config.title || 'List'}</h2>

            {config.showSearch && (
                <input
                    type="text"
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    placeholder="Search..."
                    className="w-full px-3 py-2 border rounded mb-4"
                />
            )}

            {loading && <div>Loading...</div>}

            <div className="space-y-2">
                {filteredItems.map(item => (
                    <div key={item.id} className="p-3 border rounded">
                        <h3 className="font-bold">{item.title}</h3>
                        <p className="text-gray-600">{item.description}</p>
                        <span className="text-sm">{item.status}</span>
                    </div>
                ))}
            </div>
        </div>
    );
};

export default ListComponent;`, className, name, strings.Title(strings.ReplaceAll(name, "-", " ")))
}

func generateGitignore() string {
	return `node_modules/
dist/
build/
*.log
.DS_Store
.env
.env.local
coverage/
.vscode/
.idea/
*.swp
*.swo`
}

func generateReadme(name, templateType string) string {
	return fmt.Sprintf(`# %s

A Dashspace module built with TypeScript and React.

## Development

Install dependencies:
%s

Build the module:
%s

## Publishing

Build and publish to Dashspace:
%s

## Configuration

This module can be configured through the Dashspace interface.

## Template

This module was created using the **%s** template.

## License

MIT`,
		strings.Title(strings.ReplaceAll(name, "-", " ")),
		"```bash\nnpm install\n```",
		"```bash\nnpm run build\n```",
		"```bash\ndashspace publish\n```",
		templateType)
}

func toPascalCase(s string) string {
	parts := strings.Split(s, "-")
	for i, part := range parts {
		parts[i] = strings.Title(part)
	}
	return strings.Join(parts, "")
}

func writeFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}
