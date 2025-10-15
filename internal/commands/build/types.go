package build

type DashspaceConfig struct {
	ID                    int                      `json:"id,omitempty"`
	Slug                  string                   `json:"slug,omitempty"`
	Name                  string                   `json:"name"`
	Version               string                   `json:"version"`
	Description           string                   `json:"description"`
	Author                string                   `json:"author"`
	Entry                 string                   `json:"entry"`
	Icon                  string                   `json:"icon,omitempty"`
	Category              string                   `json:"category,omitempty"`
	Tags                  []string                 `json:"tags,omitempty"`
	Permissions           []string                 `json:"permissions,omitempty"`
	RequiresSetup         bool                     `json:"requires_setup"`
	ConfigurationSteps    string                   `json:"configuration_steps"`
	Providers             []map[string]interface{} `json:"providers,omitempty"`
	ImplementedInterfaces []string                 `json:"implemented_interfaces,omitempty"`
	Checksum              string                   `json:"checksum"`
	Timestamp             string                   `json:"timestamp"`
}
