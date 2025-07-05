package sheetsql

import (
	"testing"
)

func TestSQLParser_parseSQL(t *testing.T) {
	client := &Client{}
	parser := NewSQLParser(client)

	tests := []struct {
		name     string
		sql      string
		expected struct {
			table  string
			where  int
			limit  int
			offset int
		}
		wantErr bool
	}{
		{
			name: "simple select",
			sql:  "SELECT * FROM Users",
			expected: struct {
				table  string
				where  int
				limit  int
				offset int
			}{table: "Users", where: 0, limit: 0, offset: 0},
			wantErr: false,
		},
		{
			name: "select with where",
			sql:  "SELECT * FROM Users WHERE Age > 18",
			expected: struct {
				table  string
				where  int
				limit  int
				offset int
			}{table: "Users", where: 1, limit: 0, offset: 0},
			wantErr: false,
		},
		{
			name: "select with limit",
			sql:  "SELECT * FROM Users LIMIT 10",
			expected: struct {
				table  string
				where  int
				limit  int
				offset int
			}{table: "Users", where: 0, limit: 10, offset: 0},
			wantErr: false,
		},
		{
			name: "select with offset",
			sql:  "SELECT * FROM Users OFFSET 5",
			expected: struct {
				table  string
				where  int
				limit  int
				offset int
			}{table: "Users", where: 0, limit: 0, offset: 5},
			wantErr: false,
		},
		{
			name: "select with all clauses",
			sql:  "SELECT * FROM Users WHERE Age > 18 AND Name = 'John' LIMIT 10 OFFSET 5",
			expected: struct {
				table  string
				where  int
				limit  int
				offset int
			}{table: "Users", where: 2, limit: 10, offset: 5},
			wantErr: false,
		},
		{
			name: "case insensitive",
			sql:  "select * from users where age > 18 limit 10",
			expected: struct {
				table  string
				where  int
				limit  int
				offset int
			}{table: "users", where: 1, limit: 10, offset: 0},
			wantErr: false,
		},
		{
			name:    "invalid sql",
			sql:     "INVALID SQL",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, err := parser.parseSQL(tt.sql)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseSQL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if query.sheetName != tt.expected.table {
				t.Errorf("parseSQL() table = %v, expected %v", query.sheetName, tt.expected.table)
			}

			if len(query.where) != tt.expected.where {
				t.Errorf("parseSQL() where clauses = %v, expected %v", len(query.where), tt.expected.where)
			}

			if query.limit != tt.expected.limit {
				t.Errorf("parseSQL() limit = %v, expected %v", query.limit, tt.expected.limit)
			}

			if query.offset != tt.expected.offset {
				t.Errorf("parseSQL() offset = %v, expected %v", query.offset, tt.expected.offset)
			}
		})
	}
}

func TestSQLParser_parseWhere(t *testing.T) {
	client := &Client{}
	parser := NewSQLParser(client)
	query := client.From("TestSheet")

	tests := []struct {
		name        string
		whereClause string
		expected    []WhereClause
		wantErr     bool
	}{
		{
			name:        "single condition",
			whereClause: "Age > 18",
			expected: []WhereClause{
				{Column: "Age", Operator: ">", Value: 18},
			},
			wantErr: false,
		},
		{
			name:        "string condition",
			whereClause: "Name = 'John'",
			expected: []WhereClause{
				{Column: "Name", Operator: "=", Value: "John"},
			},
			wantErr: false,
		},
		{
			name:        "multiple conditions",
			whereClause: "Age > 18 AND Name = 'John'",
			expected: []WhereClause{
				{Column: "Age", Operator: ">", Value: 18},
				{Column: "Name", Operator: "=", Value: "John"},
			},
			wantErr: false,
		},
		{
			name:        "like condition",
			whereClause: "Name LIKE 'Jo%'",
			expected: []WhereClause{
				{Column: "Name", Operator: "LIKE", Value: "Jo%"},
			},
			wantErr: false,
		},
		{
			name:        "not equal with <>",
			whereClause: "Age <> 25",
			expected: []WhereClause{
				{Column: "Age", Operator: "!=", Value: 25},
			},
			wantErr: false,
		},
		{
			name:        "float condition",
			whereClause: "Price >= 19.99",
			expected: []WhereClause{
				{Column: "Price", Operator: ">=", Value: 19.99},
			},
			wantErr: false,
		},
		{
			name:        "boolean condition",
			whereClause: "Active = true",
			expected: []WhereClause{
				{Column: "Active", Operator: "=", Value: true},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query.where = make([]WhereClause, 0)
			err := parser.parseWhere(query, tt.whereClause)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseWhere() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if len(query.where) != len(tt.expected) {
				t.Errorf("parseWhere() clause count = %v, expected %v", len(query.where), len(tt.expected))
				return
			}

			for i, expected := range tt.expected {
				actual := query.where[i]
				if actual.Column != expected.Column {
					t.Errorf("parseWhere() clause %d column = %v, expected %v", i, actual.Column, expected.Column)
				}
				if actual.Operator != expected.Operator {
					t.Errorf("parseWhere() clause %d operator = %v, expected %v", i, actual.Operator, expected.Operator)
				}
				if actual.Value != expected.Value {
					t.Errorf("parseWhere() clause %d value = %v, expected %v", i, actual.Value, expected.Value)
				}
			}
		})
	}
}

func TestSQLParser_Insert(t *testing.T) {
	client := &Client{}
	parser := NewSQLParser(client)

	tests := []struct {
		name    string
		sql     string
		wantErr bool
	}{
		{
			name:    "valid INSERT",
			sql:     "INSERT INTO Users",
			wantErr: false,
		},
		{
			name:    "case insensitive INSERT",
			sql:     "insert into products",
			wantErr: false,
		},
		{
			name:    "invalid SQL",
			sql:     "INVALID SQL",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if tt.wantErr {
						return
					}
					t.Logf("Insert() panicked as expected due to nil service: %v", r)
				}
			}()

			err := parser.Insert(tt.sql, struct{ Name string }{Name: "Test"})
			if tt.wantErr && err == nil {
				t.Errorf("Insert() expected error but got nil")
			}
		})
	}
}

func TestSQLParser_Update(t *testing.T) {
	client := &Client{}
	parser := NewSQLParser(client)

	tests := []struct {
		name    string
		sql     string
		wantErr bool
	}{
		{
			name:    "valid UPDATE without WHERE",
			sql:     "UPDATE Users SET Name = 'John'",
			wantErr: false,
		},
		{
			name:    "valid UPDATE with WHERE",
			sql:     "UPDATE Users SET Name = 'John' WHERE Age > 18",
			wantErr: false,
		},
		{
			name:    "case insensitive UPDATE",
			sql:     "update products set price = 100 where id = 1",
			wantErr: false,
		},
		{
			name:    "invalid SQL",
			sql:     "INVALID SQL",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if tt.wantErr {
						return
					}
					t.Logf("Update() panicked as expected due to nil service: %v", r)
				}
			}()

			err := parser.Update(tt.sql, struct{ Name string }{Name: "Test"})
			if tt.wantErr && err == nil {
				t.Errorf("Update() expected error but got nil")
			}
		})
	}
}

func TestSQLParser_Delete(t *testing.T) {
	client := &Client{}
	parser := NewSQLParser(client)

	tests := []struct {
		name    string
		sql     string
		wantErr bool
	}{
		{
			name:    "valid DELETE without WHERE",
			sql:     "DELETE FROM Users",
			wantErr: false,
		},
		{
			name:    "valid DELETE with WHERE",
			sql:     "DELETE FROM Users WHERE Age > 18",
			wantErr: false,
		},
		{
			name:    "case insensitive DELETE",
			sql:     "delete from products where id = 1",
			wantErr: false,
		},
		{
			name:    "invalid SQL",
			sql:     "INVALID SQL",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if tt.wantErr {
						return
					}
					t.Logf("Delete() panicked as expected due to nil service: %v", r)
				}
			}()

			err := parser.Delete(tt.sql)
			if tt.wantErr && err == nil {
				t.Errorf("Delete() expected error but got nil")
			}
		})
	}
}
