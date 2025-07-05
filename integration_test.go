package sheetsql

import (
	"context"
	"os"
	"testing"

	"google.golang.org/api/option"
)

const testSpreadsheetID = "1WoJVqoekShZGLJeoh97EVtBjuf6N7JzVSHB_l7fCT7Q"

type User struct {
	ID    int    `sheet:"ID"`
	Name  string `sheet:"Name"`
	Email string `sheet:"Email"`
	Age   int    `sheet:"Age"`
	City  string `sheet:"City"`
}

func setupIntegrationTest(t *testing.T) *Client {
	credentialsFile := os.Getenv("GOOGLE_CREDENTIALS_FILE")
	if credentialsFile == "" {
		t.Skip("GOOGLE_CREDENTIALS_FILE environment variable not set, skipping integration tests")
	}

	ctx := context.Background()
	client, err := NewClient(ctx, testSpreadsheetID, option.WithCredentialsFile(credentialsFile))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	return client
}

func TestIntegration_BasicQuery(t *testing.T) {
	client := setupIntegrationTest(t)

	var users []User
	err := client.From("Sheet1").Get(&users)
	if err != nil {
		t.Fatalf("Failed to query users: %v", err)
	}

	if len(users) == 0 {
		t.Error("Expected at least one user, got none")
	}

	t.Logf("Found %d users", len(users))
	for _, user := range users {
		t.Logf("User: %+v", user)
	}
}

func TestIntegration_WhereClause(t *testing.T) {
	client := setupIntegrationTest(t)

	var users []User
	err := client.From("Sheet1").
		Where("Age", ">", 25).
		Get(&users)
	if err != nil {
		t.Fatalf("Failed to query users with where clause: %v", err)
	}

	t.Logf("Found %d users over 25", len(users))
	for _, user := range users {
		if user.Age <= 25 {
			t.Errorf("User %s has age %d, expected > 25", user.Name, user.Age)
		}
	}
}

func TestIntegration_MultipleWhereClause(t *testing.T) {
	client := setupIntegrationTest(t)

	var users []User
	err := client.From("Sheet1").
		Where("Age", ">", 20).
		Where("City", "=", "New York").
		Get(&users)
	if err != nil {
		t.Fatalf("Failed to query users with multiple where clauses: %v", err)
	}

	t.Logf("Found %d users over 20 in New York", len(users))
	for _, user := range users {
		if user.Age <= 20 {
			t.Errorf("User %s has age %d, expected > 20", user.Name, user.Age)
		}
		if user.City != "New York" {
			t.Errorf("User %s is in %s, expected New York", user.Name, user.City)
		}
	}
}

func TestIntegration_LimitOffset(t *testing.T) {
	client := setupIntegrationTest(t)

	var allUsers []User
	err := client.From("Sheet1").Get(&allUsers)
	if err != nil {
		t.Fatalf("Failed to query all users: %v", err)
	}

	if len(allUsers) < 3 {
		t.Skip("Need at least 3 users in the sheet for limit/offset test")
	}

	var limitedUsers []User
	err = client.From("Sheet1").
		Limit(2).
		Get(&limitedUsers)
	if err != nil {
		t.Fatalf("Failed to query users with limit: %v", err)
	}

	if len(limitedUsers) != 2 {
		t.Errorf("Expected 2 users with limit, got %d", len(limitedUsers))
	}

	var offsetUsers []User
	err = client.From("Sheet1").
		Offset(1).
		Limit(2).
		Get(&offsetUsers)
	if err != nil {
		t.Fatalf("Failed to query users with offset: %v", err)
	}

	if len(offsetUsers) != 2 {
		t.Errorf("Expected 2 users with offset+limit, got %d", len(offsetUsers))
	}

	if len(allUsers) >= 2 && offsetUsers[0].ID == allUsers[0].ID {
		t.Error("Offset didn't work - first user should be different")
	}
}

func TestIntegration_SQLQuery(t *testing.T) {
	client := setupIntegrationTest(t)
	parser := NewSQLParser(client)

	var users []User
	err := parser.Query("SELECT * FROM Sheet1 WHERE Age > 25 LIMIT 3", &users)
	if err != nil {
		t.Fatalf("Failed to execute SQL query: %v", err)
	}

	if len(users) > 3 {
		t.Errorf("Expected at most 3 users, got %d", len(users))
	}

	for _, user := range users {
		if user.Age <= 25 {
			t.Errorf("User %s has age %d, expected > 25", user.Name, user.Age)
		}
	}

	t.Logf("SQL query found %d users", len(users))
}

func TestIntegration_SQLComplexQuery(t *testing.T) {
	client := setupIntegrationTest(t)
	parser := NewSQLParser(client)

	var users []User
	err := parser.Query("SELECT * FROM Sheet1 WHERE Age > 20 AND City = 'New York' LIMIT 2 OFFSET 0", &users)
	if err != nil {
		t.Fatalf("Failed to execute complex SQL query: %v", err)
	}

	if len(users) > 2 {
		t.Errorf("Expected at most 2 users, got %d", len(users))
	}

	for _, user := range users {
		if user.Age <= 20 {
			t.Errorf("User %s has age %d, expected > 20", user.Name, user.Age)
		}
		if user.City != "New York" {
			t.Errorf("User %s is in %s, expected New York", user.Name, user.City)
		}
	}

	t.Logf("Complex SQL query found %d users", len(users))
}

func TestIntegration_Insert(t *testing.T) {
	client := setupIntegrationTest(t)

	if os.Getenv("ENABLE_WRITE_TESTS") != "true" {
		t.Skip("Write tests disabled. Set ENABLE_WRITE_TESTS=true to enable")
	}

	newUser := User{
		Name:  "Test User",
		Email: "test@example.com",
		Age:   30,
		City:  "Test City",
	}

	err := client.From("Sheet1").Insert(newUser)
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}

	var users []User
	err = client.From("Sheet1").
		Where("Name", "=", "Test User").
		Get(&users)
	if err != nil {
		t.Fatalf("Failed to query inserted user: %v", err)
	}

	if len(users) == 0 {
		t.Error("Inserted user not found")
	} else {
		user := users[0]
		if user.Name != newUser.Name || user.Email != newUser.Email || user.Age != newUser.Age || user.City != newUser.City {
			t.Errorf("Inserted user data mismatch: %+v", user)
		}
	}
}