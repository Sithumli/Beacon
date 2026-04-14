package core

import (
	"encoding/json"
	"errors"
)

// Schema represents a JSON schema for input/output validation
type Schema struct {
	Type       string            `json:"type"`
	Properties map[string]Schema `json:"properties,omitempty"`
	Required   []string          `json:"required,omitempty"`
	Items      *Schema           `json:"items,omitempty"`
}

// Capability represents an agent's ability to perform a specific task
type Capability struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	InputSchema  Schema `json:"input_schema"`
	OutputSchema Schema `json:"output_schema"`
}

// Validate checks if the capability is valid
func (c *Capability) Validate() error {
	if c.Name == "" {
		return errors.New("capability name is required")
	}
	if c.Description == "" {
		return errors.New("capability description is required")
	}
	return nil
}

// MatchesInput validates that the given payload matches the input schema
// This is a basic validation - a production system would use a proper JSON schema validator
func (c *Capability) MatchesInput(payload json.RawMessage) error {
	if c.InputSchema.Type == "" {
		return nil // No schema defined, accept anything
	}

	var data interface{}
	if err := json.Unmarshal(payload, &data); err != nil {
		return err
	}

	return validateType(data, c.InputSchema)
}

func validateType(data interface{}, schema Schema) error {
	switch schema.Type {
	case "object":
		obj, ok := data.(map[string]interface{})
		if !ok {
			return errors.New("expected object type")
		}
		// Check required fields
		for _, req := range schema.Required {
			if _, exists := obj[req]; !exists {
				return errors.New("missing required field: " + req)
			}
		}
	case "string":
		if _, ok := data.(string); !ok {
			return errors.New("expected string type")
		}
	case "number":
		if _, ok := data.(float64); !ok {
			return errors.New("expected number type")
		}
	case "boolean":
		if _, ok := data.(bool); !ok {
			return errors.New("expected boolean type")
		}
	case "array":
		if _, ok := data.([]interface{}); !ok {
			return errors.New("expected array type")
		}
	}
	return nil
}
