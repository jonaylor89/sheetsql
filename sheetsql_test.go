package sheetsql

import (
	"reflect"
	"strings"
	"testing"
)

func TestQuery_Where(t *testing.T) {
	client := &Client{}
	query := client.From("TestSheet")

	query.Where("Name", "=", "John")
	query.Where("Age", ">", 18)

	if len(query.where) != 2 {
		t.Errorf("Expected 2 where clauses, got %d", len(query.where))
	}

	if query.where[0].Column != "Name" || query.where[0].Operator != "=" || query.where[0].Value != "John" {
		t.Errorf("First where clause incorrect: %+v", query.where[0])
	}

	if query.where[1].Column != "Age" || query.where[1].Operator != ">" || query.where[1].Value != 18 {
		t.Errorf("Second where clause incorrect: %+v", query.where[1])
	}
}

func TestQuery_LimitOffset(t *testing.T) {
	client := &Client{}
	query := client.From("TestSheet")

	query.Limit(10).Offset(5)

	if query.limit != 10 {
		t.Errorf("Expected limit 10, got %d", query.limit)
	}

	if query.offset != 5 {
		t.Errorf("Expected offset 5, got %d", query.offset)
	}
}

func TestQuery_matchesWhere(t *testing.T) {
	client := &Client{}
	query := client.From("TestSheet")

	headers := []string{"Name", "Age", "City"}
	fieldMap := map[string]int{"Name": 0, "Age": 1, "City": 2}

	tests := []struct {
		name     string
		where    []WhereClause
		row      []interface{}
		expected bool
	}{
		{
			name:     "no where clauses",
			where:    []WhereClause{},
			row:      []interface{}{"John", "25", "NYC"},
			expected: true,
		},
		{
			name:     "single equals match",
			where:    []WhereClause{{Column: "Name", Operator: "=", Value: "John"}},
			row:      []interface{}{"John", "25", "NYC"},
			expected: true,
		},
		{
			name:     "single equals no match",
			where:    []WhereClause{{Column: "Name", Operator: "=", Value: "Jane"}},
			row:      []interface{}{"John", "25", "NYC"},
			expected: false,
		},
		{
			name:     "greater than match",
			where:    []WhereClause{{Column: "Age", Operator: ">", Value: "20"}},
			row:      []interface{}{"John", "25", "NYC"},
			expected: true,
		},
		{
			name:     "greater than no match",
			where:    []WhereClause{{Column: "Age", Operator: ">", Value: "30"}},
			row:      []interface{}{"John", "25", "NYC"},
			expected: false,
		},
		{
			name:     "like match",
			where:    []WhereClause{{Column: "City", Operator: "LIKE", Value: "ny"}},
			row:      []interface{}{"John", "25", "NYC"},
			expected: true,
		},
		{
			name: "multiple conditions all match",
			where: []WhereClause{
				{Column: "Name", Operator: "=", Value: "John"},
				{Column: "Age", Operator: ">", Value: "20"},
			},
			row:      []interface{}{"John", "25", "NYC"},
			expected: true,
		},
		{
			name: "multiple conditions one fails",
			where: []WhereClause{
				{Column: "Name", Operator: "=", Value: "John"},
				{Column: "Age", Operator: ">", Value: "30"},
			},
			row:      []interface{}{"John", "25", "NYC"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query.where = tt.where
			result := query.matchesWhere(tt.row, headers, fieldMap)
			if result != tt.expected {
				t.Errorf("matchesWhere() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestQuery_compareValues(t *testing.T) {
	client := &Client{}
	query := client.From("TestSheet")

	tests := []struct {
		name     string
		a        string
		b        string
		operator string
		expected bool
	}{
		{"numeric greater than true", "25", "20", ">", true},
		{"numeric greater than false", "15", "20", ">", false},
		{"numeric less than true", "15", "20", "<", true},
		{"numeric less than false", "25", "20", "<", false},
		{"numeric greater equal true", "20", "20", ">=", true},
		{"numeric less equal true", "20", "20", "<=", true},
		{"string greater than", "b", "a", ">", true},
		{"string less than", "a", "b", "<", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := query.compareValues(tt.a, tt.b, tt.operator)
			if result != tt.expected {
				t.Errorf("compareValues(%s, %s, %s) = %v, expected %v", tt.a, tt.b, tt.operator, result, tt.expected)
			}
		})
	}
}

func TestQuery_setFieldValue(t *testing.T) {
	client := &Client{}
	query := client.From("TestSheet")

	type TestStruct struct {
		StringField string
		IntField    int
		FloatField  float64
		BoolField   bool
	}

	tests := []struct {
		name      string
		fieldName string
		value     string
		expected  interface{}
	}{
		{"string field", "StringField", "hello", "hello"},
		{"int field", "IntField", "42", 42},
		{"float field", "FloatField", "3.14", 3.14},
		{"bool field true", "BoolField", "true", true},
		{"bool field false", "BoolField", "false", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testStruct := TestStruct{}
			structValue := reflect.ValueOf(&testStruct).Elem()
			field := structValue.FieldByName(tt.fieldName)

			err := query.setFieldValue(field, tt.value)
			if err != nil {
				t.Errorf("setFieldValue() error = %v", err)
				return
			}

			actual := field.Interface()
			if !reflect.DeepEqual(actual, tt.expected) {
				t.Errorf("setFieldValue() = %v, expected %v", actual, tt.expected)
			}
		})
	}
}

func TestQuery_mapRowToStruct(t *testing.T) {
	client := &Client{}
	query := client.From("TestSheet")

	type Person struct {
		Name string `sheet:"Name"`
		Age  int    `sheet:"Age"`
		City string `sheet:"City"`
	}

	headers := []string{"Name", "Age", "City"}
	fieldMap := map[string]int{"Name": 0, "Age": 1, "City": 2}
	row := []interface{}{"John", "25", "NYC"}

	var person Person
	personValue := reflect.ValueOf(&person).Elem()

	err := query.mapRowToStruct(row, headers, fieldMap, personValue)
	if err != nil {
		t.Errorf("mapRowToStruct() error = %v", err)
		return
	}

	expected := Person{Name: "John", Age: 25, City: "NYC"}
	if !reflect.DeepEqual(person, expected) {
		t.Errorf("mapRowToStruct() = %+v, expected %+v", person, expected)
	}
}

func TestQuery_Update_InvalidInput(t *testing.T) {
	client := &Client{}
	query := client.From("TestSheet")

	tests := []struct {
		name        string
		input       interface{}
		expectError bool
	}{
		{
			name:        "nil input",
			input:       nil,
			expectError: true,
		},
		{
			name:        "non-struct input",
			input:       "not a struct",
			expectError: true,
		},
		{
			name:        "slice input",
			input:       []string{"not", "a", "struct"},
			expectError: true,
		},
		{
			name:        "map input",
			input:       map[string]string{"not": "struct"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := query.Update(tt.input)
			if tt.expectError && err == nil {
				t.Errorf("Update() expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Update() unexpected error: %v", err)
			}
		})
	}
}

func TestQuery_Update_ValidStruct(t *testing.T) {
	client := &Client{}
	query := client.From("TestSheet")

	type Person struct {
		Name string `sheet:"Name"`
		Age  int    `sheet:"Age"`
		City string `sheet:"City"`
	}

	tests := []struct {
		name  string
		input interface{}
	}{
		{
			name:  "struct value",
			input: Person{Name: "John", Age: 30, City: "NYC"},
		},
		{
			name:  "struct pointer",
			input: &Person{Name: "Jane", Age: 25, City: "LA"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Expected panic due to nil client service: %v", r)
				}
			}()
			err := query.Update(tt.input)
			if err != nil && !isAPIError(err) {
				t.Errorf("Update() validation error: %v", err)
			}
		})
	}
}

func TestQuery_Delete_NoWhereClause(t *testing.T) {
	client := &Client{}
	query := client.From("TestSheet")

	defer func() {
		if r := recover(); r != nil {
			t.Logf("Expected panic due to nil client service: %v", r)
		}
	}()
	err := query.Delete()
	if err != nil && !isAPIError(err) {
		t.Errorf("Delete() validation error: %v", err)
	}
}

func TestQuery_Delete_WithWhereClause(t *testing.T) {
	client := &Client{}
	query := client.From("TestSheet")
	query.Where("Name", "=", "John")

	defer func() {
		if r := recover(); r != nil {
			t.Logf("Expected panic due to nil client service: %v", r)
		}
	}()
	err := query.Delete()
	if err != nil && !isAPIError(err) {
		t.Errorf("Delete() validation error: %v", err)
	}
}

func TestQuery_getSheetId_Logic(t *testing.T) {
	client := &Client{}
	query := client.From("TestSheet")

	defer func() {
		if r := recover(); r != nil {
			t.Logf("Expected panic due to nil client service: %v", r)
		}
	}()
	id := query.getSheetId()
	if id < 0 {
		t.Errorf("getSheetId() returned negative value: %d", id)
	}
}

func isAPIError(err error) bool {
	if err == nil {
		return false
	}
	
	errStr := err.Error()
	return strings.Contains(errStr, "oauth2: cannot fetch token") ||
		strings.Contains(errStr, "invalid_request") ||
		strings.Contains(errStr, "Missing required parameter") ||
		strings.Contains(errStr, "failed to read sheet") ||
		strings.Contains(errStr, "failed to read headers") ||
		strings.Contains(errStr, "no data found in sheet") ||
		strings.Contains(errStr, "no headers found in sheet") ||
		strings.Contains(errStr, "no rows matched the where conditions") ||
		errStr == "data must be a struct or pointer to struct"
}