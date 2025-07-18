# DashSpace CLI

Official CLI for creating, developing and publishing DashSpace modules.

## 🚀 Installation

### macOS (Homebrew)

```bash
# Add DashSpace tap
brew tap dashspace/tap

# Install DashSpace CLI
brew install dashspace

# Verify installation
dashspace --version
```

### Linux (APT - Ubuntu/Debian)

```bash
# Add GPG key and repository
curl -fsSL https://packages.dashspace.dev/gpg | sudo gpg --dearmor -o /usr/share/keyrings/dashspace.gpg
echo "deb [signed-by=/usr/share/keyrings/dashspace.gpg] https://packages.dashspace.dev/apt stable main" | sudo tee /etc/apt/sources.list.d/dashspace.list

# Update and install
sudo apt update
sudo apt install dashspace-cli

# Verify installation
dashspace --version
```

### Linux (Universal Script)

```bash
# Automatic installation
curl -fsSL https://install.dashspace.dev | bash

# Or manual download
wget -O install.sh https://install.dashspace.dev
chmod +x install.sh
./install.sh
```

### Windows

```powershell
# Via Chocolatey (coming soon)
choco install dashspace-cli

# Or manual download
# Download from: https://github.com/dashspace/cli/releases/latest
```

### Manual Installation

Download the binary for your system from [GitHub Releases](https://github.com/dashspace/cli/releases/latest) and place it in your PATH.

## 🔐 Authentication

### Login

```bash
dashspace login
```

Enter your DashSpace credentials. A token will be saved in `~/.dashspace/config.json`.

### Check connection

```bash
dashspace whoami
```

### Logout

```bash
dashspace logout
```

## 📦 Creating a Module

### Interactive creation

```bash
dashspace create my-awesome-module
```

The CLI will guide you through the options:
- **Template**: React, Vanilla JS, Chart, List, Form
- **Providers**: GitHub, Slack, Asana, Notion, etc.
- **Interfaces**: ISearchable, IConfigurable, IExportable, etc.

### Creation with options

```bash
# React module with GitHub provider
dashspace create github-widget \
  --template react \
  --providers github \
  --interfaces ISearchable,IConfigurable

# Chart widget with multiple providers
dashspace create analytics-dashboard \
  --template chart \
  --providers github,asana,slack
```

### Generated structure

```
my-module/
├── devly.json          # Module manifest
├── index.js            # Main code
├── README.md           # Documentation
└── .gitignore          # Files to ignore
```

## 🛠️ Development

### Preview module

```bash
cd my-module
dashspace preview
```

This command:
1. Checks for `devly.json` file
2. Automatically opens Buildy in your browser
3. Loads your module for real-time preview

Generated URL: `http://localhost:3000/buildy?module=my-module&path=/absolute/path`

### Available templates

#### 📊 Chart Widget
Perfect for displaying charts and metrics.
```javascript
// Includes bar chart examples
// Support for different chart types
// Built-in ISearchable interface
```

#### 📋 List Widget
For displaying lists and tables.
```javascript
// Built-in search interface
// Status filtering
// Large dataset handling
```

#### 📝 Form Widget
For creating interactive forms.
```javascript
// Field validation
// API submission
// Error and success messages
```

#### ⚛️ React Custom
Empty React template for custom creations.

#### 🟨 Vanilla JS
Pure JavaScript without frameworks.

## 📤 Publishing

### Publish to store

```bash
cd my-module
dashspace publish
```

The publishing process:

1. **Validation** of manifest and code
2. **Creation** of module ZIP archive
3. **Upload** to DashSpace API
4. **Manual review** by DashSpace team
5. **Publication** to public store

### Simulation (dry-run)

```bash
dashspace publish --dry-run
```

Tests the process without actually publishing.

### Versioning

Modify the `version` field in `devly.json` before publishing:

```json
{
  "id": "my-module",
  "version": "1.1.0",
  ...
}
```

## 🔍 Store and Discovery

### Search modules

```bash
dashspace search github
dashspace search "pull requests"
dashspace search analytics
```

### Install module (future feature)

```bash
dashspace install github-pr-widget
```

## ⚙️ Configuration

### Configuration file

The CLI stores its configuration in `~/.dashspace/config.json`:

```json
{
  "api_base_url": "https://modly.dashspace.dev",
  "auth_token": "your-jwt-token",
  "username": "your-username",
  "email": "your-email@example.com"
}
```

### Environment variables

```bash
# API URL (for development)
export DASHSPACE_API_URL="http://localhost:8080"

# Debug mode
export DASHSPACE_DEBUG=true
```

## 📝 Manifest Format (devly.json)

### Complete example

```json
{
  "id": "github-pr-widget",
  "name": "GitHub Pull Requests",
  "version": "1.2.0",
  "description": "Display GitHub Pull Requests",
  "author": "My Name",
  "main": "index.js",
  
  "providers": ["github"],
  "interfaces": ["ISearchable", "IConfigurable"],
  
  "widget": {
    "type": "dashboard",
    "category": "development",
    "tags": ["github", "code-review"],
    "size": {
      "min": { "width": 3, "height": 2 },
      "default": { "width": 4, "height": 3 },
      "max": { "width": 8, "height": 6 }
    }
  },
  
  "permissions": {
    "network": ["api.github.com"]
  }
}
```

### Required fields

- `id`: Unique identifier (kebab-case)
- `name`: Display name
- `version`: Semantic version (e.g., 1.0.0)
- `description`: Short description
- `author`: Author name
- `main`: Main file (usually "index.js")

### Optional fields

- `providers`: List of required providers
- `interfaces`: Implemented interfaces
- `widget`: Widget configuration
- `permissions`: Network permissions, etc.

## 🎯 Available Interfaces

### ISearchable
Enables global search in your module.

```javascript
MyModule.search = async (query, { providers, config }) => {
  // Search in your data
  return [
    {
      id: 'result-1',
      title: 'Found result',
      description: 'Result description',
      icon: '🔍'
    }
  ];
};
```

### IConfigurable
Adds a configuration panel.

```javascript
MyModule.getConfigSchema = () => ({
  type: 'object',
  properties: {
    apiKey: {
      type: 'string',
      title: 'API Key',
      description: 'Your API key'
    },
    maxItems: {
      type: 'number',
      default: 10,
      minimum: 1,
      maximum: 100,
      title: 'Max items'
    }
  }
});
```

### IExportable
Enables data export.

```javascript
MyModule.exportData = async (format) => {
  switch (format) {
    case 'json':
      return { data: myDataObject };
    case 'csv':
      return 'col1,col2\nval1,val2';
    default:
      throw new Error('Unsupported format');
  }
};
```

## 🔧 Troubleshooting

### Common errors

**"module not found in PATH"**
```bash
# Check installation
which dashspace

# Add to PATH (Linux/macOS)
export PATH="/usr/local/bin:$PATH"
echo 'export PATH="/usr/local/bin:$PATH"' >> ~/.bashrc
```

**"permission error"**
```bash
# Reinstall with sudo
sudo rm /usr/local/bin/dashspace
curl -fsSL https://install.dashspace.dev | sudo bash
```

**"token expired"**
```bash
# Re-login
dashspace logout
dashspace login
```

**"module upload error"**
- Check your internet connection
- Ensure `devly.json` file is valid
- Verify you're logged in with `dashspace whoami`

### Debug mode

```bash
# Enable detailed logs
export DASHSPACE_DEBUG=true
dashspace publish
```

### Logs

Logs are available in:
- macOS: `~/Library/Logs/dashspace/`
- Linux: `~/.local/share/dashspace/logs/`
- Windows: `%APPDATA%/dashspace/logs/`

## 📚 Examples

### Simple GitHub module

```bash
# Create module
dashspace create github-repos --template list --providers github

# Edit code in index.js
cd github-repos
code index.js

# Test
dashspace preview

# Publish
dashspace publish
```

### Complex analytics widget

```bash
# Module with multiple providers
dashspace create team-analytics \
  --template chart \
  --providers github,asana,slack \
  --interfaces ISearchable,IConfigurable,IExportable
```

## 🆘 Support

- 📖 **Documentation**: https://docs.dashspace.dev/cli
- 💬 **Discord**: https://discord.gg/dashspace
- 🐛 **Issues**: https://github.com/dashspace/cli/issues
- 📧 **Email**: support@dashspace.dev

## 🛣️ Roadmap

- [ ] Windows Chocolatey support
- [ ] Module installation from CLI
- [ ] Community templates
- [ ] CI/CD integration (GitHub Actions)
- [ ] Advanced module scaffolding
- [ ] Automated module testing

---

**DashSpace CLI v1.0.0** - Made with ❤️ by the DashSpace team