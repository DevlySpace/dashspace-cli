package build

import (
	_ "encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
)

type Parser struct {
	moduleFile    string
	moduleContent string
}

func NewParser(moduleFile string) *Parser {
	return &Parser{
		moduleFile: moduleFile,
	}
}

func (p *Parser) loadContent() error {
	if p.moduleContent != "" {
		return nil
	}

	content, err := ioutil.ReadFile(p.moduleFile)
	if err != nil {
		return fmt.Errorf("failed to read module file: %w", err)
	}

	p.moduleContent = string(content)
	return nil
}

func (p *Parser) ExtractMetadata() (*DashspaceConfig, error) {
	if err := p.loadContent(); err != nil {
		return nil, err
	}

	config := &DashspaceConfig{
		Entry: "bundle.js",
	}

	factoryRegex := regexp.MustCompile(`(?s)function\s+DashspaceModuleFactory[^{]*\{(.*?)\n\}`)
	factoryMatch := factoryRegex.FindStringSubmatch(p.moduleContent)

	if len(factoryMatch) > 1 {
		factoryContent := factoryMatch[1]

		metadataRegex := regexp.MustCompile(`(?s)new\s+\w+\s*\([^,]+,\s*\{([^}]*)\}`)
		metadataMatch := metadataRegex.FindStringSubmatch(factoryContent)

		if len(metadataMatch) > 1 {
			metadataContent := metadataMatch[1]
			p.parseMetadataFields(metadataContent, config)
		}
	}

	if config.ID == 0 {
		return nil, fmt.Errorf("module ID not found in DashspaceModuleFactory")
	}
	if config.Name == "" {
		return nil, fmt.Errorf("module name not found in DashspaceModuleFactory")
	}
	if config.Version == "" {
		config.Version = "1.0.0"
	}
	if config.Slug == "" {
		config.Slug = sanitizeSlug(config.Name)
	}

	return config, nil
}

func (p *Parser) parseMetadataFields(content string, config *DashspaceConfig) {
	if matches := regexp.MustCompile(`id:\s*(\d+)`).FindStringSubmatch(content); len(matches) > 1 {
		config.ID, _ = strconv.Atoi(matches[1])
	}

	if matches := regexp.MustCompile(`slug:\s*['"]([^'"]+)['"]`).FindStringSubmatch(content); len(matches) > 1 {
		config.Slug = matches[1]
	}

	if matches := regexp.MustCompile(`name:\s*['"]([^'"]+)['"]`).FindStringSubmatch(content); len(matches) > 1 {
		config.Name = matches[1]
	}

	if matches := regexp.MustCompile(`version:\s*['"]([^'"]+)['"]`).FindStringSubmatch(content); len(matches) > 1 {
		config.Version = matches[1]
	}

	if matches := regexp.MustCompile(`description:\s*['"]([^'"]+)['"]`).FindStringSubmatch(content); len(matches) > 1 {
		config.Description = matches[1]
	}

	if matches := regexp.MustCompile(`author:\s*['"]([^'"]+)['"]`).FindStringSubmatch(content); len(matches) > 1 {
		config.Author = matches[1]
	}

	if matches := regexp.MustCompile(`icon:\s*['"]([^'"]+)['"]`).FindStringSubmatch(content); len(matches) > 1 {
		config.Icon = matches[1]
	}

	if matches := regexp.MustCompile(`category:\s*['"]([^'"]+)['"]`).FindStringSubmatch(content); len(matches) > 1 {
		config.Category = matches[1]
	}

	tagsRegex := regexp.MustCompile(`tags:\s*\[(.*?)\]`)
	if matches := tagsRegex.FindStringSubmatch(content); len(matches) > 1 {
		tagsContent := matches[1]
		tagRegex := regexp.MustCompile(`['"]([^'"]+)['"]`)
		tagMatches := tagRegex.FindAllStringSubmatch(tagsContent, -1)
		config.Tags = make([]string, 0)
		for _, match := range tagMatches {
			if len(match) > 1 {
				config.Tags = append(config.Tags, match[1])
			}
		}
	}
}

func (p *Parser) ExtractConfigurationSteps() ([]map[string]interface{}, error) {
	if err := p.loadContent(); err != nil {
		return nil, err
	}

	constructorRegex := regexp.MustCompile(`(?s)new\s+\w+Module\s*\([^,]+,[^}]+\}[^,]*,\s*(\w+)`)
	constructorMatch := constructorRegex.FindStringSubmatch(p.moduleContent)

	if len(constructorMatch) < 2 {
		return nil, nil
	}

	configStepsVarName := strings.TrimSpace(constructorMatch[1])

	varDefRegex := regexp.MustCompile(`(?s)(?:const|let|var)\s+` + regexp.QuoteMeta(configStepsVarName) + `[^=]*=\s*\[(.*?)\];`)
	varDefMatch := varDefRegex.FindStringSubmatch(p.moduleContent)

	if len(varDefMatch) < 2 {
		return nil, nil
	}

	extractor := &StepExtractor{}
	return extractor.ExtractSteps(varDefMatch[1])
}

func (p *Parser) ExtractProviders() ([]map[string]interface{}, error) {
	if err := p.loadContent(); err != nil {
		return nil, err
	}

	constructorRegex := regexp.MustCompile(`(?s)new\s+\w+Module\s*\([^,]+,[^}]+\}[^,]*,\s*\w+(?:,\s*(\w+))?`)
	constructorMatch := constructorRegex.FindStringSubmatch(p.moduleContent)

	if len(constructorMatch) < 2 || constructorMatch[1] == "" {
		return nil, nil
	}

	providersVarName := strings.TrimSpace(constructorMatch[1])

	varDefRegex := regexp.MustCompile(`(?s)(?:const|let|var)\s+` + regexp.QuoteMeta(providersVarName) + `[^=]*=\s*\[(.*?)\];`)
	varDefMatch := varDefRegex.FindStringSubmatch(p.moduleContent)

	if len(varDefMatch) < 2 {
		return nil, nil
	}

	extractor := &ProviderExtractor{}
	return extractor.ExtractProviders(varDefMatch[1])
}

func (p *Parser) ExtractInterfaces() ([]string, error) {
	if err := p.loadContent(); err != nil {
		return nil, err
	}

	constructorRegex := regexp.MustCompile(`(?s)new\s+\w+Module\s*\([^,]+,[^}]+\}[^,]*,\s*\w+(?:,\s*\w+)?(?:,\s*(\w+))?`)
	constructorMatch := constructorRegex.FindStringSubmatch(p.moduleContent)

	if len(constructorMatch) < 2 || constructorMatch[1] == "" {
		return nil, nil
	}

	interfacesVarName := strings.TrimSpace(constructorMatch[1])
	fmt.Printf("📋 Found interfaces variable: %s\n", interfacesVarName)

	varDefRegex := regexp.MustCompile(`(?s)(?:const|let|var)\s+` + regexp.QuoteMeta(interfacesVarName) + `[^=]*=\s*\[(.*?)\]`)
	varDefMatch := varDefRegex.FindStringSubmatch(p.moduleContent)

	if len(varDefMatch) < 2 {
		return nil, nil
	}

	interfacesContent := varDefMatch[1]
	interfaces := []string{}

	interfaceRegex := regexp.MustCompile(`ModuleInterfaces\.(\w+)`)
	interfaceMatches := interfaceRegex.FindAllStringSubmatch(interfacesContent, -1)

	for _, match := range interfaceMatches {
		if len(match) > 1 {
			interfaces = append(interfaces, match[1])
		}
	}

	if len(interfaces) > 0 {
		fmt.Printf("✅ Found %d interfaces: %v\n", len(interfaces), interfaces)

		validator := &InterfaceValidator{}
		if err := validator.ValidateImplementation(interfaces); err != nil {
			return interfaces, fmt.Errorf("interface validation failed: %w", err)
		}
	}

	return interfaces, nil
}

func (p *Parser) ExtractWebhooks() (map[string]interface{}, error) {
	if err := p.loadContent(); err != nil {
		return nil, err
	}

	// Chercher le bloc metadata complet dans le constructeur du module
	metadataRegex := regexp.MustCompile(`(?s)new\s+\w+Module\s*\(\s*context\s*,\s*\{(.*?)\}\s*,`)
	metadataMatch := metadataRegex.FindStringSubmatch(p.moduleContent)

	if len(metadataMatch) < 2 {
		return nil, nil
	}

	metadataContent := metadataMatch[1]

	// Chercher le bloc webhooks dans le metadata
	webhooksRegex := regexp.MustCompile(`(?s)webhooks:\s*\{(.*?)\}(?:\s*[,\}]|\s*$)`)
	webhooksMatch := webhooksRegex.FindStringSubmatch(metadataContent)

	if len(webhooksMatch) < 2 {
		return nil, nil
	}

	webhooksContent := webhooksMatch[1]
	webhooks := make(map[string]interface{})

	// Extract provider - support both string and Provider enum
	// Pattern: provider: Provider.GITHUB or provider: "github" or provider: 'github'
	providerRegex := regexp.MustCompile(`provider:\s*(?:Provider\.(\w+)|['"]([^'"]+)['"])`)
	if providerMatch := providerRegex.FindStringSubmatch(webhooksContent); len(providerMatch) > 1 {
		var provider string
		if providerMatch[1] != "" {
			// Provider enum (e.g., Provider.GITHUB)
			provider = strings.ToLower(providerMatch[1])
		} else if providerMatch[2] != "" {
			// String literal
			provider = providerMatch[2]
		}
		if provider != "" {
			webhooks["provider"] = provider
		}
	}

	// Extract events
	eventsRegex := regexp.MustCompile(`events:\s*\[(.*?)\]`)
	if eventsMatch := eventsRegex.FindStringSubmatch(webhooksContent); len(eventsMatch) > 1 {
		eventsContent := eventsMatch[1]
		eventRegex := regexp.MustCompile(`['"]([^'"]+)['"]`)
		eventMatches := eventRegex.FindAllStringSubmatch(eventsContent, -1)

		events := make([]string, 0)
		for _, match := range eventMatches {
			if len(match) > 1 {
				events = append(events, match[1])
			}
		}
		if len(events) > 0 {
			webhooks["events"] = events
		}
	}

	// Extract configFields
	configFieldsRegex := regexp.MustCompile(`configFields:\s*\[(.*?)\]`)
	if configFieldsMatch := configFieldsRegex.FindStringSubmatch(webhooksContent); len(configFieldsMatch) > 1 {
		configFieldsContent := configFieldsMatch[1]
		fieldRegex := regexp.MustCompile(`['"]([^'"]+)['"]`)
		fieldMatches := fieldRegex.FindAllStringSubmatch(configFieldsContent, -1)

		configFields := make([]string, 0)
		for _, match := range fieldMatches {
			if len(match) > 1 {
				configFields = append(configFields, match[1])
			}
		}
		if len(configFields) > 0 {
			webhooks["configFields"] = configFields
		}
	}

	if len(webhooks) == 0 {
		return nil, nil
	}

	return webhooks, nil
}

func (p *Parser) ExtractPermissions() ([]string, error) {
	if err := p.loadContent(); err != nil {
		return nil, err
	}

	permissions := []string{}

	permissionsRegex := regexp.MustCompile(`getPermissions\s*\(\)\s*\{[^}]*return\s*\[(.*?)\]`)
	match := permissionsRegex.FindStringSubmatch(p.moduleContent)

	if len(match) > 1 {
		permContent := match[1]
		permRegex := regexp.MustCompile(`['"]([^'"]+)['"]`)
		permMatches := permRegex.FindAllStringSubmatch(permContent, -1)

		for _, m := range permMatches {
			if len(m) > 1 {
				permissions = append(permissions, m[1])
			}
		}
	}

	return permissions, nil
}
