package importer

import (
	"testing"
)

func TestFindKeyInData(t *testing.T) {
	// Test case 1: Simple map with key at top level
	t.Run("simple map with key at top level", func(t *testing.T) {
		data := map[string]any{
			"name":  "John",
			"age":   30,
			"email": "john@example.com",
		}

		result, err := FindKeyInData(data, "name")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if result != "John" {
			t.Errorf("Expected 'John', got %v", result)
		}
	})

	// Test case 2: Nested map
	t.Run("nested map", func(t *testing.T) {
		data := map[string]any{
			"user": map[string]any{
				"profile": map[string]any{
					"name": "Jane",
					"age":  25,
				},
			},
		}

		result, err := FindKeyInData(data, "name")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if result != "Jane" {
			t.Errorf("Expected 'Jane', got %v", result)
		}
	})

	// Test case 3: Array of maps
	t.Run("array of maps", func(t *testing.T) {
		data := []any{
			map[string]any{
				"id":   1,
				"type": "user",
			},
			map[string]any{
				"id":   2,
				"name": "Bob",
			},
		}

		result, err := FindKeyInData(data, "name")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if result != "Bob" {
			t.Errorf("Expected 'Bob', got %v", result)
		}
	})

	// Test case 4: Mixed nested structure
	t.Run("mixed nested structure", func(t *testing.T) {
		data := map[string]any{
			"users": []any{
				map[string]any{
					"id": 1,
					"profile": map[string]any{
						"email": "alice@example.com",
					},
				},
				map[string]any{
					"id": 2,
					"profile": map[string]any{
						"email": "bob@example.com",
					},
				},
			},
		}

		result, err := FindKeyInData(data, "email")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if result != "alice@example.com" {
			t.Errorf("Expected 'alice@example.com', got %v", result)
		}
	})

	// Test case 5: Key not found
	t.Run("key not found", func(t *testing.T) {
		data := map[string]any{
			"name": "John",
			"age":  30,
		}

		_, err := FindKeyInData(data, "nonexistent")
		if err == nil {
			t.Error("Expected error for non-existent key")
		}

		expectedMsg := "key 'nonexistent' not found in data"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
		}
	})

	// Test case 6: Empty map
	t.Run("empty map", func(t *testing.T) {
		data := map[string]any{}

		_, err := FindKeyInData(data, "anything")
		if err == nil {
			t.Error("Expected error for empty map")
		}
	})

	// Test case 7: Empty array
	t.Run("empty array", func(t *testing.T) {
		data := []any{}

		_, err := FindKeyInData(data, "anything")
		if err == nil {
			t.Error("Expected error for empty array")
		}
	})

	// Test case 8: Primitive value (should return error)
	t.Run("primitive value", func(t *testing.T) {
		data := "just a string"

		_, err := FindKeyInData(data, "anything")
		if err == nil {
			t.Error("Expected error for primitive value")
		}
	})

	// Test case 9: Nil value
	t.Run("nil value", func(t *testing.T) {
		var data any = nil

		_, err := FindKeyInData(data, "anything")
		if err == nil {
			t.Error("Expected error for nil value")
		}
	})

	// Test case 10: Array of primitive values
	t.Run("array of primitive values", func(t *testing.T) {
		data := []any{1, 2, 3, "string", true}

		_, err := FindKeyInData(data, "anything")
		if err == nil {
			t.Error("Expected error for array of primitive values")
		}
	})

	// Test case 11: Complex nested structure with first occurrence
	t.Run("complex nested structure returns first occurrence", func(t *testing.T) {
		data := map[string]any{
			"level1": map[string]any{
				"items": []any{
					map[string]any{
						"target": "first",
					},
					map[string]any{
						"target": "second",
					},
				},
			},
		}

		result, err := FindKeyInData(data, "target")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if result != "first" {
			t.Errorf("Expected 'first', got %v", result)
		}
	})

	// Test case 12: Deeply nested structure
	t.Run("deeply nested structure", func(t *testing.T) {
		data := map[string]any{
			"a": map[string]any{
				"b": map[string]any{
					"c": map[string]any{
						"d": map[string]any{
							"e": map[string]any{
								"deep_key": "found_it",
							},
						},
					},
				},
			},
		}

		result, err := FindKeyInData(data, "deep_key")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if result != "found_it" {
			t.Errorf("Expected 'found_it', got %v", result)
		}
	})

	// Test case 13: Array containing mixed types
	t.Run("array containing mixed types", func(t *testing.T) {
		data := []any{
			"string",
			42,
			map[string]any{
				"nested_key": "value",
			},
			[]any{
				map[string]any{
					"deeper_key": "deeper_value",
				},
			},
		}

		result, err := FindKeyInData(data, "deeper_key")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if result != "deeper_value" {
			t.Errorf("Expected 'deeper_value', got %v", result)
		}
	})

	// Test case 14: Different data types as values
	t.Run("different data types as values", func(t *testing.T) {
		data := map[string]any{
			"string_val": "text",
			"int_val":    42,
			"float_val":  3.14,
			"bool_val":   true,
			"null_val":   nil,
			"array_val":  []any{1, 2, 3},
			"map_val": map[string]any{
				"nested": "value",
			},
		}

		tests := []struct {
			key      string
			expected any
		}{
			{"string_val", "text"},
			{"int_val", 42},
			{"float_val", 3.14},
			{"bool_val", true},
			{"null_val", nil},
		}

		for _, test := range tests {
			result, err := FindKeyInData(data, test.key)
			if err != nil {
				t.Errorf("Expected no error for key %s, got %v", test.key, err)
			}

			if result != test.expected {
				t.Errorf("For key %s, expected %v, got %v", test.key, test.expected, result)
			}
		}
	})
}
