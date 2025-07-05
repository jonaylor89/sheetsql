# SheetSQL

A Go ORM for Google Sheets that allows you to query and manipulate Google Sheets data using both a fluent API and SQL-like syntax.

## Features

- **Fluent API**: Chainable query builder similar to popular Go ORMs
- **SQL Support**: Write raw SQL queries against Google Sheets
- **Struct Mapping**: Map sheet rows to Go structs with tags
- **Type Safety**: Automatic type conversion and validation
- **Rich Querying**: Support for WHERE, LIMIT, OFFSET, and various operators
- **Insert Operations**: Add new rows to sheets
- **Idiomatic Go**: Follows Go best practices and conventions

## Installation

```bash
go get github.com/johannes/sheetsql
```

## Quick Start

### Setup

1. Create a Google Cloud project and enable the Google Sheets API
2. Create service account credentials and download the JSON file
3. Share your Google Sheet with the service account email

### Basic Usage

```go
package main

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
    City  string `sheet:"City"`
}

func main() {
    ctx := context.Background()
    
    // Initialize client
    client, err := sheetsql.NewClient(
        ctx, 
        "your-spreadsheet-id",
        option.WithCredentialsFile("path/to/credentials.json"),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // Query all users
    var users []User
    err = client.From("Users").Get(&users)
    if err != nil {
        log.Fatal(err)
    }
    
    for _, user := range users {
        fmt.Printf("%s (%d) - %s\n", user.Name, user.Age, user.City)
    }
}
```

## API Reference

### Fluent API

#### Basic Queries

```go
// Get all rows
var users []User
err := client.From("Users").Get(&users)

// With WHERE clause
err = client.From("Users").
    Where("Age", ">", 18).
    Get(&users)

// Multiple WHERE clauses (AND)
err = client.From("Users").
    Where("Age", ">", 18).
    Where("City", "=", "New York").
    Get(&users)

// With LIMIT and OFFSET
err = client.From("Users").
    Where("Age", ">", 18).
    Limit(10).
    Offset(5).
    Get(&users)
```

#### Supported Operators

- `=` or `==` - Equal
- `!=` - Not equal
- `>` - Greater than
- `<` - Less than
- `>=` - Greater than or equal
- `<=` - Less than or equal
- `LIKE` - Contains (case-insensitive)

#### Insert Operations

```go
newUser := User{
    Name:  "John Doe",
    Email: "john@example.com",
    Age:   30,
    City:  "New York",
}

err := client.From("Users").Insert(newUser)
```

### SQL API

For those who prefer SQL syntax:

```go
parser := sheetsql.NewSQLParser(client)

// Basic SELECT
var users []User
err := parser.Query("SELECT * FROM Users", &users)

// With WHERE clause
err = parser.Query("SELECT * FROM Users WHERE Age > 25", &users)

// Complex queries
err = parser.Query(`
    SELECT * FROM Users 
    WHERE Age > 20 AND City = 'New York' 
    LIMIT 10 OFFSET 5
`, &users)
```

#### Supported SQL Features

- `SELECT * FROM table`
- `WHERE` clauses with AND conditions
- `LIMIT` and `OFFSET`
- Operators: `=`, `!=`, `<>`, `>`, `<`, `>=`, `<=`, `LIKE`
- String literals with single or double quotes
- Automatic type conversion for numbers and booleans

### Struct Tags

Use the `sheet` tag to map struct fields to sheet columns:

```go
type User struct {
    ID       int    `sheet:"ID"`
    FullName string `sheet:"Name"`        // Maps to "Name" column
    Email    string `sheet:"Email"`
    Age      int    `sheet:"Age"`
    Location string `sheet:"City"`        // Maps to "City" column
}
```

If no tag is provided, the struct field name is used as the column name.

### Supported Types

- `string`
- `int`, `int8`, `int16`, `int32`, `int64`
- `float32`, `float64`
- `bool`

## Testing

### Unit Tests

```bash
go test ./...
```

### Integration Tests

Set up environment variables and run integration tests:

```bash
export GOOGLE_CREDENTIALS_FILE=/path/to/credentials.json
go test -v ./... -run Integration
```

### Write Tests

To enable tests that modify the sheet:

```bash
export ENABLE_WRITE_TESTS=true
go test -v ./... -run Integration
```

## Examples

Run the example program:

```bash
cd examples
export GOOGLE_CREDENTIALS_FILE=/path/to/credentials.json
go run main.go
```

Enable write examples:

```bash
export ENABLE_WRITE_EXAMPLES=true
go run main.go
```

## Sheet Structure Requirements

Your Google Sheet should have:

1. **Header row**: First row contains column names
2. **Data rows**: Subsequent rows contain data
3. **Consistent columns**: All rows should have the same number of columns

Example sheet structure:
```
| ID | Name      | Email           | Age | City      |
|----|-----------|-----------------|-----|-----------|
| 1  | John Doe  | john@email.com  | 30  | New York  |
| 2  | Jane Smith| jane@email.com  | 25  | Boston    |
```

## Authentication

### Service Account (Recommended)

1. Create a service account in Google Cloud Console
2. Download the JSON credentials file
3. Share your sheet with the service account email

```go
client, err := sheetsql.NewClient(
    ctx, 
    spreadsheetID,
    option.WithCredentialsFile("credentials.json"),
)
```

### Other Authentication Methods

```go
// With credentials JSON directly
client, err := sheetsql.NewClient(
    ctx,
    spreadsheetID,
    option.WithCredentialsJSON([]byte(credentialsJSON)),
)

// With API key (read-only, public sheets)
client, err := sheetsql.NewClient(
    ctx,
    spreadsheetID,
    option.WithAPIKey("your-api-key"),
)
```

## Error Handling

The library returns detailed errors for common issues:

```go
var users []User
err := client.From("NonExistentSheet").Get(&users)
if err != nil {
    log.Printf("Query failed: %v", err)
}
```

## Performance Considerations

- **Batch Operations**: The library fetches entire sheets and filters in memory
- **Caching**: Consider caching results for frequently accessed data
- **Sheet Size**: Performance decreases with very large sheets (>10k rows)
- **API Limits**: Google Sheets API has rate limits and quotas

## Limitations

- **Read-heavy**: Optimized for read operations
- **No Transactions**: No support for atomic operations
- **No Joins**: Cannot join data across multiple sheets
- **No Aggregations**: No built-in support for SUM, COUNT, etc.
- **Update/Delete**: Not yet implemented (coming soon)

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## License

MIT License - see LICENSE file for details.

## Roadmap

- [ ] Update and Delete operations
- [ ] Batch insert operations
- [ ] Aggregation functions
- [ ] Multiple sheet joins
- [ ] Caching layer
- [ ] Connection pooling
- [ ] Schema validation