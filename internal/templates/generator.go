package templates

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Generator struct {
	Name         string
	TemplateType string
	Providers    []string
}

func NewGenerator(name, templateType string, providers []string) *Generator {
	return &Generator{
		Name:         name,
		TemplateType: templateType,
		Providers:    providers,
	}
}

func (g *Generator) GenerateManifest() string {
	manifest := map[string]interface{}{
		"id":          g.Name,
		"name":        strings.Title(strings.ReplaceAll(g.Name, "-", " ")),
		"version":     "1.0.0",
		"description": fmt.Sprintf("Module DashSpace %s", g.Name),
		"author":      "Votre nom",
		"main":        "index.js",
		"providers":   g.Providers,
		"interfaces":  []string{"IConfigurable"},
	}

	// Ajouter ISearchable pour certains types
	if g.TemplateType == "chart" || g.TemplateType == "list" {
		interfaces := manifest["interfaces"].([]string)
		manifest["interfaces"] = append(interfaces, "ISearchable")
	}

	data, _ := json.MarshalIndent(manifest, "", "  ")
	return string(data)
}

func (g *Generator) GenerateMainFile() string {
	switch g.TemplateType {
	case "chart":
		return g.generateChartTemplate()
	case "list":
		return g.generateListTemplate()
	case "form":
		return g.generateFormTemplate()
	case "vanilla":
		return g.generateVanillaTemplate()
	default:
		return g.generateReactTemplate()
	}
}

func (g *Generator) generateReactTemplate() string {
	componentName := strings.Title(strings.ReplaceAll(g.Name, "-", ""))

	providersCode := ""
	if len(g.Providers) > 0 {
		providersCode = `
  React.useEffect(() => {
    if (providers) {
      loadData();
    }
  }, [providers]);

  const loadData = async () => {
    try {
      setLoading(true);
      // Utilisez vos providers ici
      // Exemple: const data = await providers.github.call('/user/repos');
      setLoading(false);
    } catch (error) {
      bridge.ui.notify('Erreur: ' + error.message, 'error');
      setLoading(false);
    }
  };`
	}

	return fmt.Sprintf(`const %s = ({ bridge, config = {}, providers }) => {
  const [loading, setLoading] = React.useState(false);
  const [data, setData] = React.useState([]);
%s
  return (
    <div className="p-4 bg-white rounded-lg shadow-sm border h-full">
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-lg font-bold text-gray-800">
          {config.title || '%s'}
        </h3>
        {loading && (
          <div className="w-5 h-5 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
        )}
      </div>
      
      <div className="text-center text-gray-500">
        <p className="text-lg mb-2">🚀</p>
        <p>Votre module est prêt !</p>
        <p className="text-sm mt-2">Modifiez ce code pour créer votre widget.</p>
      </div>
    </div>
  );
};

// Configuration du module
%s.getConfigSchema = () => ({
  type: 'object',
  properties: {
    title: {
      type: 'string',
      default: '%s',
      title: 'Titre du widget'
    }
  }
});

window.MyModule = %s;`, componentName, providersCode, strings.Title(strings.ReplaceAll(g.Name, "-", " ")), componentName, strings.Title(strings.ReplaceAll(g.Name, "-", " ")), componentName)
}

func (g *Generator) generateChartTemplate() string {
	componentName := strings.Title(strings.ReplaceAll(g.Name, "-", ""))

	return fmt.Sprintf(`const %s = ({ bridge, config = {}, providers }) => {
  const [data, setData] = React.useState([
    { name: 'Jan', value: 400 },
    { name: 'Fév', value: 300 },
    { name: 'Mar', value: 600 },
    { name: 'Avr', value: 800 },
    { name: 'Mai', value: 500 }
  ]);
  const [loading, setLoading] = React.useState(false);

  const refreshData = () => {
    bridge.ui.notify('Données actualisées !', 'success');
  };

  return (
    <div className="p-4 bg-gradient-to-br from-blue-50 to-indigo-100 rounded-lg border h-full">
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-lg font-bold text-gray-800">
          📊 {config.title || 'Chart Widget'}
        </h3>
        <button
          onClick={refreshData}
          className="px-3 py-1 bg-blue-500 text-white rounded text-sm hover:bg-blue-600"
        >
          🔄 Actualiser
        </button>
      </div>
      
      <div className="bg-white rounded p-4">
        <div className="flex items-end space-x-2 h-32">
          {data.map((item, index) => (
            <div key={index} className="flex-1 flex flex-col items-center">
              <div
                className="bg-blue-500 w-full rounded-t"
                style={{ height: \`\${(item.value / 800) * 100}%\` }}
              ></div>
              <span className="text-xs mt-1 text-gray-600">{item.name}</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};

%s.getConfigSchema = () => ({
  type: 'object',
  properties: {
    title: {
      type: 'string',
      default: 'Chart Widget',
      title: 'Titre du graphique'
    },
    chartType: {
      type: 'string',
      enum: ['bar', 'line', 'pie'],
      default: 'bar',
      title: 'Type de graphique'
    }
  }
});

// Interface ISearchable pour les charts
%s.search = async (query, { providers, config }) => {
  return [
    {
      id: 'chart-data',
      title: 'Données du graphique',
      description: 'Recherche dans les données du chart',
      icon: '📊'
    }
  ];
};

window.MyModule = %s;`, componentName, componentName, componentName)
}

func (g *Generator) generateListTemplate() string {
	componentName := strings.Title(strings.ReplaceAll(g.Name, "-", ""))

	return fmt.Sprintf(`const %s = ({ bridge, config = {}, providers }) => {
  const [items, setItems] = React.useState([
    { id: 1, title: 'Élément 1', status: 'actif', date: '2024-01-15' },
    { id: 2, title: 'Élément 2', status: 'en cours', date: '2024-01-14' },
    { id: 3, title: 'Élément 3', status: 'terminé', date: '2024-01-13' }
  ]);
  const [searchQuery, setSearchQuery] = React.useState('');

  const filteredItems = items.filter(item =>
    item.title.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const getStatusColor = (status) => {
    switch (status) {
      case 'actif': return 'bg-green-100 text-green-800';
      case 'en cours': return 'bg-yellow-100 text-yellow-800';
      case 'terminé': return 'bg-gray-100 text-gray-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  return (
    <div className="p-4 bg-white rounded-lg shadow-sm border h-full flex flex-col">
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-lg font-bold text-gray-800">
          📋 {config.title || 'Liste'}
        </h3>
        <span className="text-sm text-gray-500">{filteredItems.length} éléments</span>
      </div>
      
      <div className="mb-4">
        <input
          type="text"
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          placeholder="Rechercher..."
          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
        />
      </div>

      <div className="flex-1 overflow-y-auto space-y-2">
        {filteredItems.map(item => (
          <div key={item.id} className="p-3 bg-gray-50 rounded-lg border">
            <div className="flex items-center justify-between">
              <h4 className="font-medium text-gray-900">{item.title}</h4>
              <span className={\`px-2 py-1 rounded-full text-xs font-medium \${getStatusColor(item.status)}\`}>
                {item.status}
              </span>
            </div>
            <p className="text-sm text-gray-500 mt-1">{item.date}</p>
          </div>
        ))}
      </div>
    </div>
  );
};

%s.getConfigSchema = () => ({
  type: 'object',
  properties: {
    title: {
      type: 'string',
      default: 'Ma Liste',
      title: 'Titre de la liste'
    },
    maxItems: {
      type: 'number',
      default: 10,
      minimum: 5,
      maximum: 50,
      title: 'Nombre max d\'éléments'
    }
  }
});

%s.search = async (query, { providers, config }) => {
  return [
    {
      id: 'list-search',
      title: \`Recherche: \${query}\`,
      description: 'Rechercher dans la liste',
      icon: '📋'
    }
  ];
};

window.MyModule = %s;`, componentName, componentName, componentName, componentName)
}

func (g *Generator) generateFormTemplate() string {
	componentName := strings.Title(strings.ReplaceAll(g.Name, "-", ""))

	return fmt.Sprintf(`const %s = ({ bridge, config = {}, providers }) => {
  const [formData, setFormData] = React.useState({
    name: '',
    email: '',
    message: ''
  });
  const [submitting, setSubmitting] = React.useState(false);

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: value
    }));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setSubmitting(true);
    
    try {
      // Simuler un envoi
      await new Promise(resolve => setTimeout(resolve, 1000));
      
      bridge.ui.notify('Formulaire envoyé avec succès !', 'success');
      
      // Reset du formulaire
      setFormData({ name: '', email: '', message: '' });
    } catch (error) {
      bridge.ui.notify('Erreur envoi formulaire', 'error');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="p-4 bg-white rounded-lg shadow-sm border h-full">
      <h3 className="text-lg font-bold text-gray-800 mb-4">
        📝 {config.title || 'Formulaire'}
      </h3>
      
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Nom
          </label>
          <input
            type="text"
            name="name"
            value={formData.name}
            onChange={handleChange}
            required
            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Email
          </label>
          <input
            type="email"
            name="email"
            value={formData.email}
            onChange={handleChange}
            required
            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Message
          </label>
          <textarea
            name="message"
            value={formData.message}
            onChange={handleChange}
            rows={4}
            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          />
        </div>

        <button
          type="submit"
          disabled={submitting}
          className="w-full py-2 px-4 bg-blue-500 text-white rounded-md hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {submitting ? '⏳ Envoi...' : '📤 Envoyer'}
        </button>
      </form>
    </div>
  );
};

%s.getConfigSchema = () => ({
  type: 'object',
  properties: {
    title: {
      type: 'string',
      default: 'Formulaire',
      title: 'Titre du formulaire'
    },
    submitUrl: {
      type: 'string',
      title: 'URL de soumission (optionnel)'
    }
  }
});

window.MyModule = %s;`, componentName, componentName, componentName)
}

func (g *Generator) generateVanillaTemplate() string {
	componentName := strings.Title(strings.ReplaceAll(g.Name, "-", ""))

	return fmt.Sprintf(`const %s = ({ bridge, config = {}, providers }) => {
  const container = document.createElement('div');
  container.className = 'p-4 bg-white rounded-lg shadow-sm border h-full';
  
  container.innerHTML = \`
	<div class="flex items-center justify-between mb-4">
	<h3 class="text-lg font-bold text-gray-800">
	🟨 \${config.title || '%s'}
	</h3>
	<button id="refresh-btn" class="px-3 py-1 bg-blue-500 text-white rounded text-sm hover:bg-blue-600">
	🔄 Actualiser
	</button>
	</div>

	<div class="text-center text-gray-500">
	<p class="text-lg mb-2">🚀</p>
	<p>Module Vanilla JS</p>
	<p class="text-sm mt-2">Créé avec JavaScript pur</p>
	</div>

	<div id="content" class="mt-4">
	<p class="text-sm text-gray-600">
		Cliquez sur le bouton pour tester l'interactivité.
	</p>
	</div>
	\`;

  // Ajouter les event listeners
  const refreshBtn = container.querySelector('#refresh-btn');
  const content = container.querySelector('#content');
  
  refreshBtn.addEventListener('click', () => {
    bridge.ui.notify('Module actualisé !', 'success');
    content.innerHTML = \`
	<p class="text-green-600">✅ Actualisé à \${new Date().toLocaleTimeString()}</p>
	\`;
  });

  return container;
};

%s.getConfigSchema = () => ({
  type: 'object',
  properties: {
    title: {
      type: 'string',
      default: '%s',
      title: 'Titre du widget'
    },
    theme: {
      type: 'string',
      enum: ['light', 'dark'],
      default: 'light',
      title: 'Thème'
    }
  }
});

window.MyModule = %s;`, componentName, strings.Title(strings.ReplaceAll(g.Name, "-", " ")), componentName, strings.Title(strings.ReplaceAll(g.Name, "-", " ")), componentName)
}

func (g *Generator) GenerateReadme() string {
	return fmt.Sprintf(`# %s

Module DashSpace généré automatiquement.

## 🚀 Description

%s

## 📦 Installation

\`\`\`bash
	dashspace publish
	\`\`\`

## 🔧 Configuration

Ce module supporte les options de configuration suivantes :

- **title** : Titre affiché dans le widget
- Voir le schéma de configuration dans \`index.js\`

## 🔌 Providers utilisés

%s

## 🎯 Interfaces implémentées

- **IConfigurable** : Panneau de configuration
%s

## 🛠️ Développement

\`\`\`bash
	# Prévisualiser dans Buildy
	dashspace preview

	# Publier une nouvelle version
	dashspace publish
	\`\`\`

## 📄 Licence

MIT`, strings.Title(strings.ReplaceAll(g.Name, "-", " ")), fmt.Sprintf("Module DashSpace %s", g.Name), g.formatProviders(), g.formatInterfaces())
}

func (g *Generator) formatProviders() string {
	if len(g.Providers) == 0 {
		return "Aucun provider requis"
	}

	result := ""
	for _, provider := range g.Providers {
		result += fmt.Sprintf("- **%s** : Provider %s\n", provider, provider)
	}
	return result
}

func (g *Generator) formatInterfaces() string {
	if g.TemplateType == "chart" || g.TemplateType == "list" {
		return "\n- **ISearchable** : Recherche globale"
	}
	return ""
}

func (g *Generator) GenerateGitignore() string {
	return `# Dependencies
node_modules/
npm-debug.log*

# Build outputs
dist/
build/

# Environment files
.env
.env.local

# OS files
.DS_Store
Thumbs.db

# IDE files
.vscode/
.idea/
*.swp
*.swo

# Temporary files
*.tmp
*.temp

# DashSpace
.dashspace/
`
}