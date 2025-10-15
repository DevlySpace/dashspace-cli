package build

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
)

type DataSchemaExtractor struct{}

type ModuleDataSchema struct {
	ExposeData     bool                  `json:"exposeData"`
	DataType       string                `json:"dataType,omitempty"`
	Schema         *DataSchemaDefinition `json:"schema,omitempty"`
	ComputedFields []string              `json:"computedFields,omitempty"`
}

type DataSchemaDefinition struct {
	Fields       []DataFieldDefinition `json:"fields"`
	Capabilities []DataCapability      `json:"capabilities"`
}

type DataFieldDefinition struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
	Nullable    bool   `json:"nullable,omitempty"`
	Example     string `json:"example,omitempty"`
}

type DataCapability struct {
	Name   string   `json:"name"`
	Fields []string `json:"fields"`
}

func NewDataSchemaExtractor() *DataSchemaExtractor {
	return &DataSchemaExtractor{}
}

func (d *DataSchemaExtractor) ExtractDataSchema() (*ModuleDataSchema, error) {
	componentFile := findComponentFile()
	if componentFile == "" {
		hooksFile := findHooksFile()
		if hooksFile == "" {
			return &ModuleDataSchema{ExposeData: false}, nil
		}
		componentFile = hooksFile
	}

	content, err := ioutil.ReadFile(componentFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	fileContent := string(content)

	useDataProviderPattern := regexp.MustCompile(`useDataProvider\s*<[^>]*>\s*\(`)
	if !useDataProviderPattern.MatchString(fileContent) {
		useDataProviderPattern = regexp.MustCompile(`useDataProvider\s*\(`)
		if !useDataProviderPattern.MatchString(fileContent) {
			return &ModuleDataSchema{ExposeData: false}, nil
		}
	}

	fmt.Println("📊 Found useDataProvider - extracting data schema...")

	schema := &ModuleDataSchema{
		ExposeData: true,
	}

	callPattern := regexp.MustCompile(`(?s)useDataProvider\s*(?:<[^>]*>)?\s*\(\s*['"]([^'"]+)['"],\s*\{(.*?)\},\s*\{(.*?)\}`)
	callMatch := callPattern.FindStringSubmatch(fileContent)

	if len(callMatch) < 4 {
		fmt.Printf("⚠️  Warning: useDataProvider found but couldn't extract full schema\n")
		return schema, nil
	}

	payloadContent := callMatch[2]
	metadataContent := callMatch[3]

	if dataType := d.extractDataType(metadataContent); dataType != "" {
		schema.DataType = dataType
		fmt.Printf("   Data type: %s\n", dataType)
	}

	if schemaObj := d.extractSchemaObject(payloadContent); schemaObj != nil {
		schema.Schema = schemaObj
		fmt.Printf("   Fields: %d\n", len(schemaObj.Fields))
		fmt.Printf("   Capabilities: %d\n", len(schemaObj.Capabilities))
	}

	if computedFields := d.extractComputedFields(payloadContent); len(computedFields) > 0 {
		schema.ComputedFields = computedFields
		fmt.Printf("   Computed metrics: %d\n", len(computedFields))
	}

	return schema, nil
}

func (d *DataSchemaExtractor) extractDataType(metadataContent string) string {
	typePattern := regexp.MustCompile(`type:\s*['"]([^'"]+)['"]`)
	if match := typePattern.FindStringSubmatch(metadataContent); len(match) > 1 {
		return match[1]
	}
	return ""
}

func (d *DataSchemaExtractor) extractSchemaObject(payloadContent string) *DataSchemaDefinition {
	schemaPattern := regexp.MustCompile(`(?s)schema:\s*\{(.*?)\}(?:\s*,|\s*\})`)
	schemaMatch := schemaPattern.FindStringSubmatch(payloadContent)

	if len(schemaMatch) < 2 {
		return nil
	}

	schemaContent := schemaMatch[1]

	schemaObj := &DataSchemaDefinition{}

	fieldsPattern := regexp.MustCompile(`(?s)fields:\s*\[(.*?)\]`)
	if fieldsMatch := fieldsPattern.FindStringSubmatch(schemaContent); len(fieldsMatch) > 1 {
		schemaObj.Fields = d.extractFields(fieldsMatch[1])
	}

	capabilitiesPattern := regexp.MustCompile(`(?s)capabilities:\s*\[(.*?)\]`)
	if capMatch := capabilitiesPattern.FindStringSubmatch(schemaContent); len(capMatch) > 1 {
		schemaObj.Capabilities = d.extractCapabilities(capMatch[1])
	}

	return schemaObj
}

func (d *DataSchemaExtractor) extractFields(fieldsContent string) []DataFieldDefinition {
	var fields []DataFieldDefinition

	fieldPattern := regexp.MustCompile(`\{[^}]+\}`)
	fieldMatches := fieldPattern.FindAllString(fieldsContent, -1)

	for _, fieldStr := range fieldMatches {
		field := DataFieldDefinition{}

		if match := regexp.MustCompile(`name:\s*['"]([^'"]+)['"]`).FindStringSubmatch(fieldStr); len(match) > 1 {
			field.Name = match[1]
		}

		if match := regexp.MustCompile(`type:\s*['"]([^'"]+)['"]`).FindStringSubmatch(fieldStr); len(match) > 1 {
			field.Type = match[1]
		}

		if match := regexp.MustCompile(`description:\s*['"]([^'"]+)['"]`).FindStringSubmatch(fieldStr); len(match) > 1 {
			field.Description = match[1]
		}

		if match := regexp.MustCompile(`nullable:\s*(true|false)`).FindStringSubmatch(fieldStr); len(match) > 1 {
			field.Nullable = match[1] == "true"
		}

		if match := regexp.MustCompile(`example:\s*['"]?([^'",}]+)['"]?`).FindStringSubmatch(fieldStr); len(match) > 1 {
			field.Example = strings.TrimSpace(match[1])
		}

		if field.Name != "" && field.Type != "" {
			fields = append(fields, field)
		}
	}

	return fields
}

func (d *DataSchemaExtractor) extractCapabilities(capContent string) []DataCapability {
	var capabilities []DataCapability

	capPattern := regexp.MustCompile(`\{[^}]+\}`)
	capMatches := capPattern.FindAllString(capContent, -1)

	for _, capStr := range capMatches {
		capability := DataCapability{}

		if match := regexp.MustCompile(`name:\s*['"]([^'"]+)['"]`).FindStringSubmatch(capStr); len(match) > 1 {
			capability.Name = match[1]
		}

		fieldsPattern := regexp.MustCompile(`fields:\s*\[(.*?)\]`)
		if fieldsMatch := fieldsPattern.FindStringSubmatch(capStr); len(fieldsMatch) > 1 {
			fieldsList := fieldsMatch[1]
			fieldNamePattern := regexp.MustCompile(`['"]([^'"]+)['"]`)
			fieldNames := fieldNamePattern.FindAllStringSubmatch(fieldsList, -1)

			for _, fn := range fieldNames {
				if len(fn) > 1 {
					capability.Fields = append(capability.Fields, fn[1])
				}
			}
		}

		if capability.Name != "" {
			capabilities = append(capabilities, capability)
		}
	}

	return capabilities
}

func (d *DataSchemaExtractor) extractComputedFields(payloadContent string) []string {
	var computedFields []string

	computedPattern := regexp.MustCompile(`(?s)computed:\s*\{(.*?)\}`)
	if computedMatch := computedPattern.FindStringSubmatch(payloadContent); len(computedMatch) > 1 {
		computedContent := computedMatch[1]

		fieldPattern := regexp.MustCompile(`([a-zA-Z_][a-zA-Z0-9_]*)\s*:`)
		fieldMatches := fieldPattern.FindAllStringSubmatch(computedContent, -1)

		for _, match := range fieldMatches {
			if len(match) > 1 {
				computedFields = append(computedFields, match[1])
			}
		}
	}

	return computedFields
}

func (d *DataSchemaExtractor) ValidateDataSchema(schema *ModuleDataSchema) error {
	if !schema.ExposeData {
		return nil
	}

	fmt.Println("🔍 Validating data schema...")

	if schema.DataType == "" {
		fmt.Printf("⚠️  Warning: No data type specified (using 'generic')\n")
	} else {
		validTypes := map[string]bool{
			"issue-tracker":     true,
			"payment-system":    true,
			"error-tracking":    true,
			"deployment-system": true,
			"task-management":   true,
			"communication":     true,
			"analytics":         true,
			"generic":           true,
		}

		if !validTypes[schema.DataType] {
			return fmt.Errorf("invalid data type: %s", schema.DataType)
		}
	}

	if schema.Schema != nil {
		if len(schema.Schema.Fields) == 0 {
			fmt.Printf("⚠️  Warning: Schema defined but no fields found\n")
		}

		for _, field := range schema.Schema.Fields {
			validFieldTypes := map[string]bool{
				"string": true, "number": true, "date": true,
				"boolean": true, "array": true, "object": true,
			}

			if !validFieldTypes[field.Type] {
				return fmt.Errorf("invalid field type '%s' for field '%s'", field.Type, field.Name)
			}
		}

		for _, cap := range schema.Schema.Capabilities {
			validCapabilities := map[string]bool{
				"filterable": true, "sortable": true, "groupable": true,
				"aggregatable": true, "time-series": true,
			}

			if !validCapabilities[cap.Name] {
				fmt.Printf("⚠️  Warning: Unknown capability '%s'\n", cap.Name)
			}
		}
	}

	fmt.Println("✅ Data schema validated")
	return nil
}

func (d *DataSchemaExtractor) SerializeToJSON(schema *ModuleDataSchema) (string, error) {
	if schema == nil || !schema.ExposeData {
		return "", nil
	}

	jsonData, err := json.Marshal(schema)
	if err != nil {
		return "", fmt.Errorf("failed to serialize data schema: %w", err)
	}

	return string(jsonData), nil
}

func findHooksFile() string {
	candidates := []string{
		"hooks/useModuleData.ts",
		"hooks/useData.ts",
		"src/hooks/useModuleData.ts",
		"src/hooks/useData.ts",
		"hooks/index.ts",
		"src/hooks/index.ts",
	}

	for _, candidate := range candidates {
		if fileExists(candidate) {
			return candidate
		}
	}

	return ""
}
