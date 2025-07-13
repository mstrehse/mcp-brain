package actions

import (
	"fmt"
	"strings"

	"github.com/mstrehse/mcp-brain/pkg/contracts"
)

// validateTemplate validates a task template
func validateTemplate(template *contracts.TaskTemplate) error {
	if template.Name == "" {
		return fmt.Errorf("template name is required")
	}
	if template.Description == "" {
		return fmt.Errorf("template description is required")
	}
	if len(template.Tasks) == 0 {
		return fmt.Errorf("template must have at least one task")
	}

	// Validate parameters
	for paramName, param := range template.Parameters {
		if paramName == "" {
			return fmt.Errorf("parameter name cannot be empty")
		}
		if param.Type == "" {
			return fmt.Errorf("parameter '%s' must have a type", paramName)
		}
		if param.Type == "enum" && len(param.Values) == 0 {
			return fmt.Errorf("enum parameter '%s' must have values", paramName)
		}
	}

	return nil
}

// validateTemplateParameters validates that provided parameters match template requirements
func validateTemplateParameters(template *contracts.TaskTemplate, parameters map[string]string) error {
	// Check required parameters
	for paramName, param := range template.Parameters {
		if param.Required {
			if value, exists := parameters[paramName]; !exists || value == "" {
				return fmt.Errorf("required parameter '%s' is missing", paramName)
			}
		}
	}

	// Validate enum parameters
	for paramName, value := range parameters {
		if param, exists := template.Parameters[paramName]; exists {
			if param.Type == "enum" && len(param.Values) > 0 {
				valid := false
				for _, validValue := range param.Values {
					if value == validValue {
						valid = true
						break
					}
				}
				if !valid {
					return fmt.Errorf("parameter '%s' must be one of: %s", paramName, strings.Join(param.Values, ", "))
				}
			}
		}
	}

	return nil
}
