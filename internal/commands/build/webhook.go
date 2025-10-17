package build

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
)

type WebhookValidator struct{}

func NewWebhookValidator() *WebhookValidator {
	return &WebhookValidator{}
}

// ValidateWebhookConfiguration validates webhook configuration in Module.ts metadata
func (v *WebhookValidator) ValidateWebhookConfiguration(webhooks map[string]interface{}) error {
	if webhooks == nil || len(webhooks) == 0 {
		return nil // No webhooks to validate
	}

	fmt.Println("🔍 Validating webhook configuration...")

	// Validate provider (required)
	provider, ok := webhooks["provider"].(string)
	if !ok || provider == "" {
		return fmt.Errorf("webhook provider is required in metadata.webhooks")
	}

	// Validate it's using Provider enum
	if !strings.HasPrefix(provider, "Provider.") && !v.isValidProviderName(provider) {
		return fmt.Errorf("webhook provider should use Provider enum (e.g., Provider.GITHUB) or valid provider name, got: %s", provider)
	}

	// Validate events (required)
	events, ok := webhooks["events"].([]string)
	if !ok || len(events) == 0 {
		return fmt.Errorf("webhook events array is required and cannot be empty")
	}

	// Validate event format
	for _, event := range events {
		if event == "" {
			return fmt.Errorf("empty event found in webhooks.events")
		}
		// Events should be lowercase with underscores (e.g., "issues", "pull_request")
		if !v.isValidEventFormat(event) {
			fmt.Printf("⚠️  Warning: event '%s' should use lowercase with underscores (e.g., 'pull_request', 'issues')\n", event)
		}
	}

	// Validate configFields (required for GitHub-like providers)
	configFields, ok := webhooks["configFields"].([]string)
	if !ok {
		configFields, ok = webhooks["config_fields"].([]string)
	}

	if ok && len(configFields) > 0 {
		for _, field := range configFields {
			if field == "" {
				return fmt.Errorf("empty config field found in webhooks.configFields")
			}
		}
		fmt.Printf("✅ Found %d webhook config fields: %v\n", len(configFields), configFields)
	} else {
		fmt.Printf("⚠️  Warning: No configFields specified. Webhooks may need configuration fields like ['owner', 'repo']\n")
	}

	// Check for optional handler
	if handler, ok := webhooks["handler"]; ok {
		if handler == nil {
			fmt.Printf("⚠️  Warning: webhook handler is null in metadata\n")
		}
	}

	fmt.Println("✅ Webhook configuration validated")
	return nil
}

// ValidateWebhookImplementation checks Module.ts implementation
func (v *WebhookValidator) ValidateWebhookImplementation(webhooks map[string]interface{}) error {
	if webhooks == nil || len(webhooks) == 0 {
		return nil
	}

	moduleFile := findModuleFile()
	if moduleFile == "" {
		return fmt.Errorf("Module.ts not found")
	}

	content, err := ioutil.ReadFile(moduleFile)
	if err != nil {
		return fmt.Errorf("failed to read module file: %w", err)
	}

	moduleContent := string(content)

	fmt.Println("🔍 Validating webhook implementation in Module.ts...")

	// Check if module extends BaseModule
	if !strings.Contains(moduleContent, "extends BaseModule") {
		return fmt.Errorf("Module must extend BaseModule to use webhooks")
	}

	// Check for registerWebhookHandler calls
	registerPattern := regexp.MustCompile(`this\.registerWebhookHandler\(['"]([^'"]+)['"]`)
	registerMatches := registerPattern.FindAllStringSubmatch(moduleContent, -1)

	if len(registerMatches) == 0 {
		fmt.Printf("⚠️  Warning: No registerWebhookHandler calls found in initialize(). Webhooks may not be handled.\n")
	} else {
		fmt.Printf("✅ Found %d webhook handler registrations\n", len(registerMatches))

		// Validate registered events match metadata
		events := webhooks["events"].([]string)
		registeredEvents := make(map[string]bool)

		for _, match := range registerMatches {
			if len(match) > 1 {
				registeredEvents[match[1]] = true
			}
		}

		for _, event := range events {
			if !registeredEvents[event] {
				fmt.Printf("⚠️  Warning: Event '%s' declared in metadata but no handler registered in initialize()\n", event)
			}
		}
	}

	// Check for handler methods
	handlerPattern := regexp.MustCompile(`(?:private|protected|async)\s+(?:async\s+)?handle\w+Event\s*\([^)]*WebhookEvent[^)]*\)`)
	if !handlerPattern.MatchString(moduleContent) {
		fmt.Printf("⚠️  Warning: No webhook event handler methods found (e.g., handleIssuesEvent)\n")
	}

	// Check for WebhookEvent import
	if !strings.Contains(moduleContent, "WebhookEvent") {
		return fmt.Errorf("WebhookEvent type not imported from dashspace-lib")
	}

	// Validate handler signature pattern
	events := webhooks["events"].([]string)
	for _, event := range events {
		// Convert event name to method name (issues -> Issues, pull_request -> PullRequest)
		methodName := v.eventToMethodName(event)
		handlerMethodPattern := regexp.MustCompile(fmt.Sprintf(`handle%sEvent\s*\([^)]*(?:event|webhook)[^)]*:\s*WebhookEvent`, methodName))

		if !handlerMethodPattern.MatchString(moduleContent) {
			fmt.Printf("⚠️  Warning: Expected handler method 'handle%sEvent(event: WebhookEvent)' not found\n", methodName)
		}
	}

	fmt.Println("✅ Webhook implementation validated")
	return nil
}

// ValidateComponentWebhookUsage checks if Component.tsx uses webhook hooks
func (v *WebhookValidator) ValidateComponentWebhookUsage(webhooks map[string]interface{}) error {
	if webhooks == nil || len(webhooks) == 0 {
		return nil
	}

	componentFile := findComponentFile()
	if componentFile == "" {
		fmt.Printf("⚠️  Warning: Component.tsx not found, skipping component webhook validation\n")
		return nil
	}

	content, err := ioutil.ReadFile(componentFile)
	if err != nil {
		return nil
	}

	componentContent := string(content)

	fmt.Println("🔍 Validating webhook usage in Component.tsx...")

	// Check for webhook hook import
	if !strings.Contains(componentContent, "useWebhookEvents") && !strings.Contains(componentContent, "useWebhook") {
		fmt.Printf("⚠️  Warning: No webhook hooks (useWebhookEvents/useWebhook) imported in Component\n")
		return nil
	}

	// Check for webhook hook usage
	webhookHookPattern := regexp.MustCompile(`use(?:Webhook|WebhookEvents)\(`)
	if !webhookHookPattern.MatchString(componentContent) {
		fmt.Printf("⚠️  Warning: Webhook hooks imported but not used in Component\n")
	}

	// Check for custom webhook hook (e.g., useGitHubIssueWebhook)
	customHookPattern := regexp.MustCompile(`use\w+Webhook\(\)`)
	customHookMatches := customHookPattern.FindAllString(componentContent, -1)

	if len(customHookMatches) > 0 {
		fmt.Printf("✅ Found custom webhook hook usage: %v\n", customHookMatches)
	}

	// Check for webhook event handling
	events := webhooks["events"].([]string)
	for _, event := range events {
		// Look for event in useWebhookEvents or custom handlers
		if strings.Contains(componentContent, fmt.Sprintf("'%s", event)) ||
			strings.Contains(componentContent, fmt.Sprintf("\"%s", event)) {
			// Event is referenced, good
		} else {
			fmt.Printf("⚠️  Warning: Event '%s' not explicitly handled in Component\n", event)
		}
	}

	fmt.Println("✅ Component webhook usage validated")
	return nil
}

// ValidateWebhookSecurity checks for security best practices
func (v *WebhookValidator) ValidateWebhookSecurity() error {
	fmt.Println("🔍 Checking webhook security best practices...")

	moduleFile := findModuleFile()
	if moduleFile == "" {
		return nil
	}

	content, err := ioutil.ReadFile(moduleFile)
	if err != nil {
		return nil
	}

	moduleContent := string(content)
	warnings := []string{}

	// Check for error handling in webhook handlers
	handlerPattern := regexp.MustCompile(`(?s)handle\w+Event\s*\([^)]+\)\s*\{([^}]+)\}`)
	handlers := handlerPattern.FindAllStringSubmatch(moduleContent, -1)

	for _, handler := range handlers {
		if len(handler) > 1 {
			handlerBody := handler[1]
			if !strings.Contains(handlerBody, "try") && !strings.Contains(handlerBody, "catch") {
				warnings = append(warnings, "Webhook handlers should use try/catch for error handling")
				break
			}
		}
	}

	// Check for data validation
	if !strings.Contains(moduleContent, "event.data") {
		warnings = append(warnings, "Webhook handlers should validate event.data before processing")
	}

	// Check for emit on webhook events (best practice for reactivity)
	if !strings.Contains(moduleContent, "this.emit") {
		warnings = append(warnings, "Consider using this.emit() to notify Component of webhook events")
	}

	if len(warnings) > 0 {
		fmt.Println("⚠️  Webhook security recommendations:")
		for _, warning := range warnings {
			fmt.Printf("   - %s\n", warning)
		}
	} else {
		fmt.Println("✅ Webhook security checks passed")
	}

	return nil
}

// Helper functions

func (v *WebhookValidator) isValidProviderName(provider string) bool {
	validProviders := map[string]bool{
		"github":   true,
		"gitlab":   true,
		"google":   true,
		"slack":    true,
		"stripe":   true,
		"twilio":   true,
		"discord":  true,
		"linear":   true,
		"notion":   true,
		"airtable": true,
	}
	return validProviders[strings.ToLower(provider)]
}

func (v *WebhookValidator) isValidEventFormat(event string) bool {
	// Events should be lowercase with optional underscores (e.g., "issues", "pull_request")
	validFormat := regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	return validFormat.MatchString(event)
}

func (v *WebhookValidator) eventToMethodName(event string) string {
	// Convert "issues" -> "Issues", "pull_request" -> "PullRequest"
	parts := strings.Split(event, "_")
	result := ""
	for _, part := range parts {
		if len(part) > 0 {
			result += strings.ToUpper(part[0:1]) + part[1:]
		}
	}
	return result
}
