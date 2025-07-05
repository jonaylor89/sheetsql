package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/johannes/sheetsql"
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

func main() {
	credentialsFile := os.Getenv("GOOGLE_CREDENTIALS_FILE")
	if credentialsFile == "" {
		log.Fatal("GOOGLE_CREDENTIALS_FILE environment variable not set")
	}

	ctx := context.Background()
	client, err := sheetsql.NewClient(ctx, testSpreadsheetID, option.WithCredentialsFile(credentialsFile))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	fmt.Println("=== SheetSQL Examples ===")

	// Example 1: Basic query
	fmt.Println("1. Basic Query - Get all users:")
	var allUsers []User
	err = client.From("Sheet1").Get(&allUsers)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		for _, user := range allUsers {
			fmt.Printf("   %d: %s (%s) - Age: %d, City: %s\n", user.ID, user.Name, user.Email, user.Age, user.City)
		}
	}

	// Example 2: Query with WHERE clause
	fmt.Println("\n2. Query with WHERE clause - Users over 25:")
	var adultUsers []User
	err = client.From("Sheet1").
		Where("Age", ">", 25).
		Get(&adultUsers)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		for _, user := range adultUsers {
			fmt.Printf("   %s - Age: %d\n", user.Name, user.Age)
		}
	}

	// Example 3: Multiple WHERE clauses
	fmt.Println("\n3. Multiple WHERE clauses - Users over 20 in New York:")
	var nyUsers []User
	err = client.From("Sheet1").
		Where("Age", ">", 20).
		Where("City", "=", "New York").
		Get(&nyUsers)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		for _, user := range nyUsers {
			fmt.Printf("   %s - Age: %d, City: %s\n", user.Name, user.Age, user.City)
		}
	}

	// Example 4: LIKE operator
	fmt.Println("\n4. LIKE operator - Users with 'John' in name:")
	var johnUsers []User
	err = client.From("Sheet1").
		Where("Name", "LIKE", "John").
		Get(&johnUsers)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		for _, user := range johnUsers {
			fmt.Printf("   %s\n", user.Name)
		}
	}

	// Example 5: Limit and Offset
	fmt.Println("\n5. Limit and Offset - First 2 users:")
	var limitedUsers []User
	err = client.From("Sheet1").
		Limit(2).
		Get(&limitedUsers)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		for _, user := range limitedUsers {
			fmt.Printf("   %s\n", user.Name)
		}
	}

	fmt.Println("\n   Next 2 users (with offset):")
	var offsetUsers []User
	err = client.From("Sheet1").
		Limit(2).
		Offset(2).
		Get(&offsetUsers)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		for _, user := range offsetUsers {
			fmt.Printf("   %s\n", user.Name)
		}
	}

	// Example 6: SQL Queries
	fmt.Println("\n6. SQL Query - SELECT * FROM Users WHERE Age > 25 LIMIT 3:")
	parser := sheetsql.NewSQLParser(client)
	var sqlUsers []User
	err = parser.Query("SELECT * FROM Sheet1 WHERE Age > 25 LIMIT 3", &sqlUsers)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		for _, user := range sqlUsers {
			fmt.Printf("   %s - Age: %d\n", user.Name, user.Age)
		}
	}

	// Example 7: Complex SQL Query
	fmt.Println("\n7. Complex SQL Query - SELECT * FROM Users WHERE Age > 20 AND City = 'New York':")
	var complexSQLUsers []User
	err = parser.Query("SELECT * FROM Sheet1 WHERE Age > 20 AND City = 'New York'", &complexSQLUsers)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		for _, user := range complexSQLUsers {
			fmt.Printf("   %s - Age: %d, City: %s\n", user.Name, user.Age, user.City)
		}
	}

	// Example 8: Insert (commented out by default to avoid modifying the sheet)
	if os.Getenv("ENABLE_WRITE_EXAMPLES") == "true" {
		fmt.Println("\n8. Insert new user:")
		newUser := User{
			Name:  "Example User",
			Email: "example@test.com",
			Age:   28,
			City:  "Example City",
		}

		err = client.From("Sheet1").Insert(newUser)
		if err != nil {
			log.Printf("Error inserting user: %v", err)
		} else {
			fmt.Printf("   Successfully inserted: %s\n", newUser.Name)
		}
	} else {
		fmt.Println("\n8. Insert example (disabled - set ENABLE_WRITE_EXAMPLES=true to enable)")
	}

	fmt.Println("\n=== Examples Complete ===")
}
