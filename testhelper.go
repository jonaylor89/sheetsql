package sheetsql

import (
	"context"
	"fmt"
	"testing"

	"google.golang.org/api/option"
)

type MockSheetsService struct {
	mockData map[string][][]interface{}
}

func NewMockSheetsService() *MockSheetsService {
	return &MockSheetsService{
		mockData: make(map[string][][]interface{}),
	}
}

func (m *MockSheetsService) AddSheetData(sheetName string, data [][]interface{}) {
	m.mockData[sheetName] = data
}

func (m *MockSheetsService) GetSheetData(sheetName string) [][]interface{} {
	if data, exists := m.mockData[sheetName]; exists {
		return data
	}
	return nil
}

func NewMockClient(spreadsheetID string) *Client {
	return &Client{
		service:      nil,
		spreadsheetID: spreadsheetID,
	}
}

func SetupTestData() *MockSheetsService {
	mock := NewMockSheetsService()
	
	testUsers := [][]interface{}{
		{"ID", "Name", "Email", "Age", "City"},
		{1, "John Doe", "john@example.com", 30, "New York"},
		{2, "Jane Smith", "jane@example.com", 25, "Los Angeles"},
		{3, "Bob Johnson", "bob@example.com", 35, "Chicago"},
		{4, "Alice Brown", "alice@example.com", 28, "New York"},
		{5, "Charlie Wilson", "charlie@example.com", 22, "Boston"},
	}
	
	mock.AddSheetData("Users", testUsers)
	
	return mock
}

func TestNewClient(t *testing.T) {
	client := NewMockClient("test-spreadsheet-id")
	if client == nil {
		t.Fatal("NewMockClient returned nil")
	}
	
	if client.spreadsheetID != "test-spreadsheet-id" {
		t.Errorf("Expected spreadsheet ID 'test-spreadsheet-id', got %s", client.spreadsheetID)
	}
}

func TestClientCreation(t *testing.T) {
	ctx := context.Background()
	
	client, err := NewClient(ctx, "test-id")
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	
	if client.spreadsheetID != "test-id" {
		t.Errorf("Expected spreadsheet ID 'test-id', got %s", client.spreadsheetID)
	}
	
	if client.service == nil {
		t.Error("Expected service to be initialized")
	}
}

func TestClientCreationWithOptions(t *testing.T) {
	ctx := context.Background()
	
	_, err := NewClient(ctx, "test-id", option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("NewClient with options failed: %v", err)
	}
}

func BenchmarkQuery_matchesWhere(b *testing.B) {
	client := &Client{}
	query := client.From("TestSheet")
	
	query.Where("Name", "=", "John")
	query.Where("Age", ">", 25)
	
	headers := []string{"Name", "Age", "City"}
	fieldMap := map[string]int{"Name": 0, "Age": 1, "City": 2}
	row := []interface{}{"John", "30", "NYC"}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		query.matchesWhere(row, headers, fieldMap)
	}
}

func BenchmarkQuery_compareValues(b *testing.B) {
	client := &Client{}
	query := client.From("TestSheet")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		query.compareValues("30", "25", ">")
	}
}

func ExampleMockSheetsService() {
	mock := NewMockSheetsService()
	
	testData := [][]interface{}{
		{"Name", "Age"},
		{"John", 30},
		{"Jane", 25},
	}
	
	mock.AddSheetData("TestSheet", testData)
	
	data := mock.GetSheetData("TestSheet")
	fmt.Printf("Retrieved %d rows from TestSheet\n", len(data))
	
	// Output: Retrieved 3 rows from TestSheet
}

func TestSheetDataStructure(t *testing.T) {
	mock := SetupTestData()
	
	users := mock.GetSheetData("Users")
	if len(users) != 6 {
		t.Errorf("Expected 6 rows (including header), got %d", len(users))
	}
	
	if fmt.Sprintf("%v", users[0]) != "[ID Name Email Age City]" {
		t.Errorf("Header row incorrect: %v", users[0])
	}
	
	if fmt.Sprintf("%v", users[1]) != "[1 John Doe john@example.com 30 New York]" {
		t.Errorf("First data row incorrect: %v", users[1])
	}
}