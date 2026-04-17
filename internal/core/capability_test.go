package core

import (
	"encoding/json"
	"testing"
)

func TestCapabilityValidate(t *testing.T) {
	tests := []struct {
		name       string
		capability Capability
		wantErr    bool
		errMsg     string
	}{
		{
			name: "valid capability",
			capability: Capability{
				Name:        "echo",
				Description: "Echo capability that returns the input",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			capability: Capability{
				Description: "Echo capability",
			},
			wantErr: true,
			errMsg:  "capability name is required",
		},
		{
			name: "missing description",
			capability: Capability{
				Name: "echo",
			},
			wantErr: true,
			errMsg:  "capability description is required",
		},
		{
			name:       "empty capability",
			capability: Capability{},
			wantErr:    true,
			errMsg:     "capability name is required",
		},
		{
			name: "with input schema",
			capability: Capability{
				Name:        "process",
				Description: "Process data",
				InputSchema: Schema{
					Type: "object",
					Properties: map[string]Schema{
						"data": {Type: "string"},
					},
					Required: []string{"data"},
				},
			},
			wantErr: false,
		},
		{
			name: "with output schema",
			capability: Capability{
				Name:        "compute",
				Description: "Compute result",
				OutputSchema: Schema{
					Type: "number",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.capability.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != nil && err.Error() != tt.errMsg {
				t.Errorf("Validate() error message = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestCapabilityMatchesInput(t *testing.T) {
	tests := []struct {
		name       string
		capability Capability
		payload    string
		wantErr    bool
	}{
		{
			name: "no schema - accepts anything",
			capability: Capability{
				Name:        "echo",
				Description: "Echo",
			},
			payload: `{"any": "data"}`,
			wantErr: false,
		},
		{
			name: "no schema - accepts string",
			capability: Capability{
				Name:        "echo",
				Description: "Echo",
			},
			payload: `"hello"`,
			wantErr: false,
		},
		{
			name: "object schema - valid object",
			capability: Capability{
				Name:        "process",
				Description: "Process",
				InputSchema: Schema{
					Type: "object",
					Properties: map[string]Schema{
						"name": {Type: "string"},
					},
					Required: []string{"name"},
				},
			},
			payload: `{"name": "test"}`,
			wantErr: false,
		},
		{
			name: "object schema - missing required field",
			capability: Capability{
				Name:        "process",
				Description: "Process",
				InputSchema: Schema{
					Type: "object",
					Properties: map[string]Schema{
						"name": {Type: "string"},
					},
					Required: []string{"name"},
				},
			},
			payload: `{"other": "value"}`,
			wantErr: true,
		},
		{
			name: "object schema - wrong type (array instead of object)",
			capability: Capability{
				Name:        "process",
				Description: "Process",
				InputSchema: Schema{
					Type: "object",
				},
			},
			payload: `[1, 2, 3]`,
			wantErr: true,
		},
		{
			name: "string schema - valid string",
			capability: Capability{
				Name:        "greet",
				Description: "Greet",
				InputSchema: Schema{
					Type: "string",
				},
			},
			payload: `"hello world"`,
			wantErr: false,
		},
		{
			name: "string schema - wrong type (number)",
			capability: Capability{
				Name:        "greet",
				Description: "Greet",
				InputSchema: Schema{
					Type: "string",
				},
			},
			payload: `123`,
			wantErr: true,
		},
		{
			name: "number schema - valid number",
			capability: Capability{
				Name:        "compute",
				Description: "Compute",
				InputSchema: Schema{
					Type: "number",
				},
			},
			payload: `42.5`,
			wantErr: false,
		},
		{
			name: "number schema - valid integer",
			capability: Capability{
				Name:        "compute",
				Description: "Compute",
				InputSchema: Schema{
					Type: "number",
				},
			},
			payload: `100`,
			wantErr: false,
		},
		{
			name: "number schema - wrong type (string)",
			capability: Capability{
				Name:        "compute",
				Description: "Compute",
				InputSchema: Schema{
					Type: "number",
				},
			},
			payload: `"not a number"`,
			wantErr: true,
		},
		{
			name: "boolean schema - valid true",
			capability: Capability{
				Name:        "toggle",
				Description: "Toggle",
				InputSchema: Schema{
					Type: "boolean",
				},
			},
			payload: `true`,
			wantErr: false,
		},
		{
			name: "boolean schema - valid false",
			capability: Capability{
				Name:        "toggle",
				Description: "Toggle",
				InputSchema: Schema{
					Type: "boolean",
				},
			},
			payload: `false`,
			wantErr: false,
		},
		{
			name: "boolean schema - wrong type",
			capability: Capability{
				Name:        "toggle",
				Description: "Toggle",
				InputSchema: Schema{
					Type: "boolean",
				},
			},
			payload: `"true"`,
			wantErr: true,
		},
		{
			name: "array schema - valid array",
			capability: Capability{
				Name:        "batch",
				Description: "Batch process",
				InputSchema: Schema{
					Type: "array",
				},
			},
			payload: `[1, 2, 3]`,
			wantErr: false,
		},
		{
			name: "array schema - empty array",
			capability: Capability{
				Name:        "batch",
				Description: "Batch process",
				InputSchema: Schema{
					Type: "array",
				},
			},
			payload: `[]`,
			wantErr: false,
		},
		{
			name: "array schema - wrong type",
			capability: Capability{
				Name:        "batch",
				Description: "Batch process",
				InputSchema: Schema{
					Type: "array",
				},
			},
			payload: `{"not": "array"}`,
			wantErr: true,
		},
		{
			name: "invalid JSON",
			capability: Capability{
				Name:        "echo",
				Description: "Echo",
				InputSchema: Schema{
					Type: "object",
				},
			},
			payload: `{invalid json}`,
			wantErr: true,
		},
		{
			name: "object with multiple required fields - all present",
			capability: Capability{
				Name:        "create",
				Description: "Create resource",
				InputSchema: Schema{
					Type: "object",
					Properties: map[string]Schema{
						"name":  {Type: "string"},
						"email": {Type: "string"},
						"age":   {Type: "number"},
					},
					Required: []string{"name", "email"},
				},
			},
			payload: `{"name": "John", "email": "john@example.com"}`,
			wantErr: false,
		},
		{
			name: "object with multiple required fields - one missing",
			capability: Capability{
				Name:        "create",
				Description: "Create resource",
				InputSchema: Schema{
					Type: "object",
					Properties: map[string]Schema{
						"name":  {Type: "string"},
						"email": {Type: "string"},
					},
					Required: []string{"name", "email"},
				},
			},
			payload: `{"name": "John"}`,
			wantErr: true,
		},
		{
			name: "null payload",
			capability: Capability{
				Name:        "echo",
				Description: "Echo",
				InputSchema: Schema{
					Type: "object",
				},
			},
			payload: `null`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.capability.MatchesInput(json.RawMessage(tt.payload))
			if (err != nil) != tt.wantErr {
				t.Errorf("MatchesInput() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateType(t *testing.T) {
	tests := []struct {
		name    string
		data    interface{}
		schema  Schema
		wantErr bool
	}{
		{
			name:    "object type - valid",
			data:    map[string]interface{}{"key": "value"},
			schema:  Schema{Type: "object"},
			wantErr: false,
		},
		{
			name:    "object type - with required field present",
			data:    map[string]interface{}{"name": "test"},
			schema:  Schema{Type: "object", Required: []string{"name"}},
			wantErr: false,
		},
		{
			name:    "object type - with required field missing",
			data:    map[string]interface{}{"other": "value"},
			schema:  Schema{Type: "object", Required: []string{"name"}},
			wantErr: true,
		},
		{
			name:    "string type - valid",
			data:    "hello",
			schema:  Schema{Type: "string"},
			wantErr: false,
		},
		{
			name:    "string type - invalid",
			data:    123,
			schema:  Schema{Type: "string"},
			wantErr: true,
		},
		{
			name:    "number type - valid float",
			data:    float64(42.5),
			schema:  Schema{Type: "number"},
			wantErr: false,
		},
		{
			name:    "number type - invalid",
			data:    "42",
			schema:  Schema{Type: "number"},
			wantErr: true,
		},
		{
			name:    "boolean type - valid true",
			data:    true,
			schema:  Schema{Type: "boolean"},
			wantErr: false,
		},
		{
			name:    "boolean type - invalid",
			data:    "true",
			schema:  Schema{Type: "boolean"},
			wantErr: true,
		},
		{
			name:    "array type - valid",
			data:    []interface{}{1, 2, 3},
			schema:  Schema{Type: "array"},
			wantErr: false,
		},
		{
			name:    "array type - invalid",
			data:    "not an array",
			schema:  Schema{Type: "array"},
			wantErr: true,
		},
		{
			name:    "unknown type - no validation",
			data:    "anything",
			schema:  Schema{Type: "unknown"},
			wantErr: false,
		},
		{
			name:    "empty type - no validation",
			data:    map[string]interface{}{"any": "thing"},
			schema:  Schema{Type: ""},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateType(tt.data, tt.schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateType() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSchemaJSON(t *testing.T) {
	schema := Schema{
		Type: "object",
		Properties: map[string]Schema{
			"name": {Type: "string"},
			"age":  {Type: "number"},
			"tags": {
				Type:  "array",
				Items: &Schema{Type: "string"},
			},
		},
		Required: []string{"name"},
	}

	// Test marshaling
	data, err := json.Marshal(schema)
	if err != nil {
		t.Fatalf("failed to marshal schema: %v", err)
	}

	// Test unmarshaling
	var decoded Schema
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal schema: %v", err)
	}

	if decoded.Type != schema.Type {
		t.Errorf("Type mismatch: got %s, want %s", decoded.Type, schema.Type)
	}
	if len(decoded.Properties) != len(schema.Properties) {
		t.Errorf("Properties count mismatch: got %d, want %d", len(decoded.Properties), len(schema.Properties))
	}
	if len(decoded.Required) != len(schema.Required) {
		t.Errorf("Required count mismatch: got %d, want %d", len(decoded.Required), len(schema.Required))
	}
	// Check that nested Items in tags property was preserved
	tagsSchema, ok := decoded.Properties["tags"]
	if !ok {
		t.Error("tags property should exist")
	} else if tagsSchema.Items == nil {
		t.Error("tags.Items should not be nil")
	} else if tagsSchema.Items.Type != "string" {
		t.Errorf("tags.Items.Type = %s, want string", tagsSchema.Items.Type)
	}
}
