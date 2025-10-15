package build

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type StepExtractor struct{}

func (e *StepExtractor) ExtractSteps(stepsContent string) ([]map[string]interface{}, error) {
	var configSteps []map[string]interface{}

	allStepStarts := regexp.MustCompile(`new\s+ConfigurationStep\s*\(`).FindAllStringIndex(stepsContent, -1)

	for i, stepStart := range allStepStarts {
		startPos := stepStart[0]
		endPos := len(stepsContent)

		if i < len(allStepStarts)-1 {
			endPos = allStepStarts[i+1][0]
		}

		stepBlock := stepsContent[startPos:endPos]

		openBrace := strings.Index(stepBlock, "{")
		if openBrace == -1 {
			continue
		}

		braceCount := 0
		closeBrace := -1
		for j := openBrace; j < len(stepBlock); j++ {
			if stepBlock[j] == '{' {
				braceCount++
			} else if stepBlock[j] == '}' {
				braceCount--
				if braceCount == 0 {
					closeBrace = j
					break
				}
			}
		}

		if closeBrace == -1 {
			continue
		}

		stepContent := stepBlock[openBrace+1 : closeBrace]
		step := e.parseStep(stepContent)
		if step != nil && len(step) > 0 {
			configSteps = append(configSteps, step)
		}
	}

	return configSteps, nil
}

func (e *StepExtractor) parseStep(stepContent string) map[string]interface{} {
	step := make(map[string]interface{})

	if id := extractStringValue(stepContent, "id"); id != "" {
		step["id"] = id
	}
	if title := extractStringValue(stepContent, "title"); title != "" {
		step["title"] = title
	}
	if description := extractStringValue(stepContent, "description"); description != "" {
		step["description"] = description
	}
	if order := extractNumberValue(stepContent, "order"); order != -1 {
		step["order"] = order
	}
	if optional := extractBoolValue(stepContent, "optional"); optional {
		step["optional"] = true
	}

	fields := e.extractFields(stepContent)
	if len(fields) > 0 {
		step["fields"] = fields
	}

	return step
}

func (e *StepExtractor) extractFields(stepContent string) []map[string]interface{} {
	var fields []map[string]interface{}

	fieldsIdx := strings.Index(stepContent, "fields:")
	if fieldsIdx == -1 {
		return fields
	}

	remaining := stepContent[fieldsIdx+7:]
	arrayStart := strings.Index(remaining, "[")
	if arrayStart == -1 {
		return fields
	}

	actualStart := fieldsIdx + 7 + arrayStart

	bracketCount := 0
	arrayEnd := -1
	for i := actualStart; i < len(stepContent); i++ {
		if stepContent[i] == '[' {
			bracketCount++
		} else if stepContent[i] == ']' {
			bracketCount--
			if bracketCount == 0 {
				arrayEnd = i
				break
			}
		}
	}

	if arrayEnd == -1 {
		return fields
	}

	fieldsContent := stepContent[actualStart+1 : arrayEnd]

	allFieldStarts := regexp.MustCompile(`new\s+(\w+Field)\s*\(`).FindAllStringIndex(fieldsContent, -1)

	for i, fieldStart := range allFieldStarts {
		startPos := fieldStart[0]
		endPos := len(fieldsContent)

		if i < len(allFieldStarts)-1 {
			endPos = allFieldStarts[i+1][0]
		}

		fieldBlock := fieldsContent[startPos:endPos]

		fieldTypeMatch := regexp.MustCompile(`new\s+(\w+Field)`).FindStringSubmatch(fieldBlock)
		if len(fieldTypeMatch) < 2 {
			continue
		}
		fieldType := fieldTypeMatch[1]

		openBrace := strings.Index(fieldBlock, "{")
		if openBrace == -1 {
			continue
		}

		braceCount := 0
		closeBrace := -1
		for j := openBrace; j < len(fieldBlock); j++ {
			if fieldBlock[j] == '{' {
				braceCount++
			} else if fieldBlock[j] == '}' {
				braceCount--
				if braceCount == 0 {
					closeBrace = j
					break
				}
			}
		}

		if closeBrace == -1 {
			continue
		}

		fieldContent := fieldBlock[openBrace+1 : closeBrace]

		field := map[string]interface{}{
			"type": getFieldType(fieldType),
		}

		if name := extractStringValue(fieldContent, "name"); name != "" {
			field["name"] = name
		}
		if label := extractStringValue(fieldContent, "label"); label != "" {
			field["label"] = label
		}
		if description := extractStringValue(fieldContent, "description"); description != "" {
			field["description"] = description
		}
		if placeholder := extractStringValue(fieldContent, "placeholder"); placeholder != "" {
			field["placeholder"] = placeholder
		}
		if defaultValue := extractDefaultValue(fieldContent); defaultValue != nil {
			field["defaultValue"] = defaultValue
		}

		validation := e.extractValidation(fieldContent, fieldType)
		if len(validation) > 0 {
			field["validation"] = validation
		}

		fields = append(fields, field)
	}

	return fields
}

func (e *StepExtractor) extractValidation(fieldContent string, fieldType string) map[string]interface{} {
	validation := make(map[string]interface{})
	hasValidation := false

	if validationContent := extractValidationBlock(fieldContent); validationContent != "" {
		if required := extractBoolValue(validationContent, "required"); required {
			validation["required"] = true
			hasValidation = true
		}
		if pattern := extractStringValue(validationContent, "pattern"); pattern != "" {
			validation["pattern"] = pattern
			hasValidation = true
		}
		if customMessage := extractStringValue(validationContent, "customMessage"); customMessage != "" {
			validation["customMessage"] = customMessage
			hasValidation = true
		}
	}

	if fieldType == "NumberField" {
		if min := extractNumberValue(fieldContent, "min"); min != -1 {
			validation["min"] = min
			hasValidation = true
		}
		if max := extractNumberValue(fieldContent, "max"); max != -1 {
			validation["max"] = max
			hasValidation = true
		}
	}

	if fieldType == "SelectField" {
		if options := extractSelectOptions(fieldContent); len(options) > 0 {
			validation["options"] = options
			hasValidation = true
		}
	}

	if hasValidation {
		return validation
	}
	return nil
}

type ProviderExtractor struct{}

func (p *ProviderExtractor) ExtractProviders(providersContent string) ([]map[string]interface{}, error) {
	var providers []map[string]interface{}

	objectRegex := regexp.MustCompile(`\{[^{}]*(?:\{[^{}]*\}[^{}]*)?\}`)
	objectMatches := objectRegex.FindAllString(providersContent, -1)

	for _, objectStr := range objectMatches {
		provider := make(map[string]interface{})

		providerNameMatch := regexp.MustCompile(`provider:\s*Provider\.(\w+)`).FindStringSubmatch(objectStr)
		if len(providerNameMatch) >= 2 {
			provider["name"] = strings.ToLower(providerNameMatch[1])
		} else {
			continue
		}

		if regexp.MustCompile(`required:\s*true`).MatchString(objectStr) {
			provider["required"] = true
		} else if regexp.MustCompile(`required:\s*false`).MatchString(objectStr) {
			provider["required"] = false
		} else {
			provider["required"] = true
		}

		scopesRegex := regexp.MustCompile(`scopes:\s*\[(.*?)\]`)
		if scopesMatch := scopesRegex.FindStringSubmatch(objectStr); len(scopesMatch) > 1 {
			scopesContent := scopesMatch[1]
			scopeRegex := regexp.MustCompile(`['"]([^'"]+)['"]`)
			scopeMatches := scopeRegex.FindAllStringSubmatch(scopesContent, -1)

			scopes := make([]string, 0)
			for _, scopeMatch := range scopeMatches {
				if len(scopeMatch) > 1 {
					scopes = append(scopes, scopeMatch[1])
				}
			}
			if len(scopes) > 0 {
				provider["scopes"] = scopes
			}
		}

		descRegex := regexp.MustCompile(`description:\s*['"]([^'"]+)['"]`)
		if descMatch := descRegex.FindStringSubmatch(objectStr); len(descMatch) > 1 {
			provider["description"] = descMatch[1]
		}

		providers = append(providers, provider)
	}

	return providers, nil
}

func getFieldType(fieldClass string) string {
	switch fieldClass {
	case "TextField":
		return "text"
	case "NumberField":
		return "number"
	case "SelectField":
		return "select"
	case "BooleanField":
		return "boolean"
	case "PasswordField":
		return "password"
	case "UrlField":
		return "url"
	case "DateField":
		return "date"
	case "ColorField":
		return "color"
	case "EmailField":
		return "email"
	default:
		return "text"
	}
}

func extractValidationBlock(content string) string {
	validationIdx := strings.Index(content, "validation:")
	if validationIdx == -1 {
		return ""
	}

	startIdx := strings.Index(content[validationIdx:], "{")
	if startIdx == -1 {
		return ""
	}
	startIdx += validationIdx

	braceCount := 0
	endIdx := startIdx
	for i := startIdx; i < len(content); i++ {
		if content[i] == '{' {
			braceCount++
		} else if content[i] == '}' {
			braceCount--
			if braceCount == 0 {
				endIdx = i
				break
			}
		}
	}

	if endIdx <= startIdx {
		return ""
	}

	return content[startIdx+1 : endIdx]
}

func extractSelectOptions(content string) []map[string]interface{} {
	var options []map[string]interface{}

	optionsIdx := strings.Index(content, "options:")
	if optionsIdx == -1 {
		return options
	}

	startIdx := strings.Index(content[optionsIdx:], "[")
	if startIdx == -1 {
		return options
	}
	startIdx += optionsIdx

	bracketCount := 0
	endIdx := startIdx
	for i := startIdx; i < len(content); i++ {
		if content[i] == '[' {
			bracketCount++
		} else if content[i] == ']' {
			bracketCount--
			if bracketCount == 0 {
				endIdx = i
				break
			}
		}
	}

	if endIdx <= startIdx {
		return options
	}

	optionsContent := content[startIdx+1 : endIdx]

	optionRegex := regexp.MustCompile(`\{\s*value:\s*['"]([^'"]+)['"],\s*label:\s*['"]([^'"]+)['"]\s*\}`)
	optionMatches := optionRegex.FindAllStringSubmatch(optionsContent, -1)

	for _, match := range optionMatches {
		if len(match) >= 3 {
			options = append(options, map[string]interface{}{
				"value": match[1],
				"label": match[2],
			})
		}
	}

	return options
}

func extractStringValue(content, key string) string {
	pattern := fmt.Sprintf(`%s:\s*['"]([^'"]+)['"]`, key)
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func extractNumberValue(content, key string) int {
	pattern := fmt.Sprintf(`%s:\s*(\d+)`, key)
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		if num, err := strconv.Atoi(matches[1]); err == nil {
			return num
		}
	}
	return -1
}

func extractBoolValue(content, key string) bool {
	pattern := fmt.Sprintf(`%s:\s*(true|false)`, key)
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return matches[1] == "true"
	}
	return false
}

func extractDefaultValue(content string) interface{} {
	stringPattern := `defaultValue:\s*['"]([^'"]+)['"]`
	if re := regexp.MustCompile(stringPattern); re.MatchString(content) {
		matches := re.FindStringSubmatch(content)
		if len(matches) > 1 {
			return matches[1]
		}
	}

	numberPattern := `defaultValue:\s*(\d+)`
	if re := regexp.MustCompile(numberPattern); re.MatchString(content) {
		matches := re.FindStringSubmatch(content)
		if len(matches) > 1 {
			if num, err := strconv.Atoi(matches[1]); err == nil {
				return num
			}
		}
	}

	boolPattern := `defaultValue:\s*(true|false)`
	if re := regexp.MustCompile(boolPattern); re.MatchString(content) {
		matches := re.FindStringSubmatch(content)
		if len(matches) > 1 {
			return matches[1] == "true"
		}
	}

	return nil
}
