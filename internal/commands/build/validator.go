package build

import (
	"fmt"
)

type Validator struct{}

func NewValidator() *Validator {
	return &Validator{}
}

func (v *Validator) ValidateStructure() error {
	if findModuleFile() == "" {
		return fmt.Errorf("Module.ts or Module.tsx not found")
	}

	if findComponentFile() == "" {
		fmt.Printf("⚠️  Warning: Component.tsx not found - module may not have a UI component\n")
	}

	if !fileExists("package.json") {
		return fmt.Errorf("package.json not found")
	}

	return nil
}

func (v *Validator) ValidateMetadata(config *DashspaceConfig) error {
	if config.ID == 0 {
		return fmt.Errorf("module ID is required")
	}

	if config.Name == "" {
		return fmt.Errorf("module name is required")
	}

	if config.Version == "" {
		return fmt.Errorf("module version is required")
	}

	if config.Slug == "" {
		return fmt.Errorf("module slug is required")
	}

	return nil
}

func (v *Validator) ValidatePermissions(permissions []string) error {
	validPermissions := map[string]bool{
		"storage:read":     true,
		"storage:write":    true,
		"ui:notifications": true,
		"ui:modals":        true,
		"network:external": true,
	}

	for _, perm := range permissions {
		if !validPermissions[perm] {
			fmt.Printf("⚠️  Warning: Unknown permission '%s'\n", perm)
		}
	}

	return nil
}
