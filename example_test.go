package sheetsql_test

import (
	"context"
	"fmt"
	"log"

	"github.com/johannes/sheetsql"
	"google.golang.org/api/option"
)

type User struct {
	ID    int    `sheet:"ID"`
	Name  string `sheet:"Name"`
	Email string `sheet:"Email"`
	Age   int    `sheet:"Age"`
}

func ExampleClient_fluent() {
	ctx := context.Background()
	
	client, err := sheetsql.NewClient(ctx, "your-spreadsheet-id", option.WithCredentialsFile("credentials.json"))
	if err != nil {
		log.Fatal(err)
	}

	var users []User
	err = client.From("Users").
		Where("Age", ">", 18).
		Where("Name", "LIKE", "John").
		Limit(10).
		Get(&users)

	if err != nil {
		log.Fatal(err)
	}

	for _, user := range users {
		fmt.Printf("User: %s (%s), Age: %d\n", user.Name, user.Email, user.Age)
	}
}

func ExampleClient_sql() {
	ctx := context.Background()
	
	client, err := sheetsql.NewClient(ctx, "your-spreadsheet-id", option.WithCredentialsFile("credentials.json"))
	if err != nil {
		log.Fatal(err)
	}

	parser := sheetsql.NewSQLParser(client)
	
	var users []User
	err = parser.Query("SELECT * FROM Users WHERE Age > 18 AND Name LIKE 'John' LIMIT 10", &users)
	if err != nil {
		log.Fatal(err)
	}

	for _, user := range users {
		fmt.Printf("User: %s (%s), Age: %d\n", user.Name, user.Email, user.Age)
	}
}

func ExampleClient_insert() {
	ctx := context.Background()
	
	client, err := sheetsql.NewClient(ctx, "your-spreadsheet-id", option.WithCredentialsFile("credentials.json"))
	if err != nil {
		log.Fatal(err)
	}

	newUser := User{
		Name:  "Jane Doe",
		Email: "jane@example.com",
		Age:   25,
	}

	err = client.From("Users").Insert(newUser)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("User inserted successfully")
}