package build

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
)

type InterfaceValidator struct{}

func (v *InterfaceValidator) ValidateImplementation(declaredInterfaces []string) error {
	componentFile := findComponentFile()
	if componentFile == "" {
		return fmt.Errorf("Component.tsx not found")
	}

	content, err := ioutil.ReadFile(componentFile)
	if err != nil {
		return fmt.Errorf("failed to read component file: %w", err)
	}

	componentContent := string(content)

	if !strings.Contains(componentContent, "useModuleInterfaces") {
		return fmt.Errorf("Component does not use useModuleInterfaces hook")
	}

	handlersRegex := regexp.MustCompile(`(?s)(?:const|let|var)\s+handlers\s*(?::\s*\w+\s*)?=\s*\{(.*?)\}\s*satisfies\s+InterfaceHandlers`)
	handlersMatch := handlersRegex.FindStringSubmatch(componentContent)

	if len(handlersMatch) < 2 {
		handlersRegex = regexp.MustCompile(`(?s)(?:const|let|var)\s+handlers\s*(?::\s*\w+\s*)?=\s*\{(.*?)\}(?:\s*;|\s*\n|\s*$)`)
		handlersMatch = handlersRegex.FindStringSubmatch(componentContent)
	}

	if len(handlersMatch) < 2 {
		return fmt.Errorf("handlers object not found in Component")
	}

	handlersContent := handlersMatch[1]
	implementedInterfaces := v.extractImplementedInterfaces(handlersContent)

	for _, declared := range declaredInterfaces {
		found := false
		for _, implemented := range implementedInterfaces {
			if declared == implemented || "I"+declared == implemented || declared == "I"+implemented {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("interface %s declared in Module.ts but not implemented in Component.tsx", declared)
		}
	}

	for _, interfaceName := range implementedInterfaces {
		if err := v.validateInterfaceMethods(handlersContent, interfaceName); err != nil {
			return fmt.Errorf("interface %s validation failed: %w", interfaceName, err)
		}
	}

	fmt.Printf("✅ All %d interfaces properly implemented\n", len(declaredInterfaces))
	return nil
}

func (v *InterfaceValidator) extractImplementedInterfaces(handlersContent string) []string {
	interfaces := []string{}

	interfaceRegex := regexp.MustCompile(`(?m)^\s*["']?(\w+)["']?\s*:\s*\{`)
	matches := interfaceRegex.FindAllStringSubmatch(handlersContent, -1)

	for _, match := range matches {
		if len(match) > 1 {
			interfaceName := match[1]
			if strings.HasPrefix(interfaceName, "I") {
				interfaces = append(interfaces, interfaceName)
			}
		}
	}

	return interfaces
}

func (v *InterfaceValidator) validateInterfaceMethods(handlersContent string, interfaceName string) error {
	requiredMethods := map[string][]string{
		"ISearchable":   {"search", "getSearchResults", "getSearchFilters", "translateUQLQuery", "getUQLCapabilities"},
		"IRefreshable":  {"refresh", "getLastRefresh", "setAutoRefresh"},
		"IExportable":   {"export", "getSupportedFormats", "exportData"},
		"IFilterable":   {"applyFilter", "clearFilters", "getCurrentFilter", "translateUQLQuery", "getUQLCapabilities"},
		"IThemeable":    {"applyTheme", "getThemeConfig", "getSupportedThemes"},
		"INotifiable":   {"sendNotification", "subscribe", "unsubscribe"},
		"IDataProvider": {"getData", "getDataSchema", "subscribe"},
		"ISchedulable":  {"schedule", "unschedule", "getSchedule"},
	}

	methods, exists := requiredMethods[interfaceName]
	if !exists {
		fmt.Printf("⚠️  Warning: Unknown interface %s, skipping method validation\n", interfaceName)
		return nil
	}

	escapedName := regexp.QuoteMeta(interfaceName)
	interfacePattern := fmt.Sprintf(`(?s)%s\s*:\s*\{(.*?)\}\s*satisfies\s+%s`, escapedName, escapedName)
	interfaceBlockRegex := regexp.MustCompile(interfacePattern)
	blockMatch := interfaceBlockRegex.FindStringSubmatch(handlersContent)

	if len(blockMatch) < 2 {
		interfacePattern = fmt.Sprintf(`(?s)%s\s*:\s*\{(.*?)\}(?:\s*,|\s*\n)`, escapedName)
		interfaceBlockRegex = regexp.MustCompile(interfacePattern)
		blockMatch = interfaceBlockRegex.FindStringSubmatch(handlersContent)
	}

	if len(blockMatch) < 2 {
		return fmt.Errorf("could not find interface block for %s", interfaceName)
	}

	interfaceBlock := blockMatch[1]

	missingMethods := []string{}
	for _, method := range methods {
		methodPattern := fmt.Sprintf(`\b%s\s*[:=]\s*(?:async\s+)?(?:\([^)]*\)\s*=>|\w)`, regexp.QuoteMeta(method))
		methodRegex := regexp.MustCompile(methodPattern)
		if !methodRegex.MatchString(interfaceBlock) {
			missingMethods = append(missingMethods, method)
		}
	}

	if len(missingMethods) > 0 {
		return fmt.Errorf("missing methods: %v", missingMethods)
	}

	return nil
}

func isKnownInterface(name string) bool {
	knownInterfaces := []string{
		"Searchable", "Refreshable", "Exportable", "Filterable",
		"Configurable", "Themeable", "Notifiable", "DataProvider",
		"Widget", "Schedulable",
	}

	for _, known := range knownInterfaces {
		if name == known {
			return true
		}
	}
	return false
}

func findComponentFile() string {
	candidates := []string{
		"Component.tsx",
		"src/Component.tsx",
		"Component.ts",
		"src/Component.ts",
	}

	for _, candidate := range candidates {
		if fileExists(candidate) {
			return candidate
		}
	}

	return ""
}
