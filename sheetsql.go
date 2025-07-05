package sheetsql

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type Client struct {
	service      *sheets.Service
	spreadsheetID string
}

type Query struct {
	client    *Client
	sheetName string
	where     []WhereClause
	limit     int
	offset    int
}

type WhereClause struct {
	Column   string
	Operator string
	Value    interface{}
}

func NewClient(ctx context.Context, spreadsheetID string, opts ...option.ClientOption) (*Client, error) {
	srv, err := sheets.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create sheets service: %w", err)
	}

	return &Client{
		service:      srv,
		spreadsheetID: spreadsheetID,
	}, nil
}

func (c *Client) From(sheetName string) *Query {
	return &Query{
		client:    c,
		sheetName: sheetName,
		where:     make([]WhereClause, 0),
	}
}

func (q *Query) Where(column, operator string, value interface{}) *Query {
	q.where = append(q.where, WhereClause{
		Column:   column,
		Operator: operator,
		Value:    value,
	})
	return q
}

func (q *Query) Limit(limit int) *Query {
	q.limit = limit
	return q
}

func (q *Query) Offset(offset int) *Query {
	q.offset = offset
	return q
}

func (q *Query) Get(dest interface{}) error {
	destValue := reflect.ValueOf(dest)
	if destValue.Kind() != reflect.Ptr || destValue.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("dest must be a pointer to a slice")
	}

	sliceValue := destValue.Elem()
	elemType := sliceValue.Type().Elem()

	readRange := fmt.Sprintf("%s!A:Z", q.sheetName)
	resp, err := q.client.service.Spreadsheets.Values.Get(q.client.spreadsheetID, readRange).Do()
	if err != nil {
		return fmt.Errorf("failed to read sheet: %w", err)
	}

	if len(resp.Values) == 0 {
		return nil
	}

	headers := make([]string, len(resp.Values[0]))
	for i, header := range resp.Values[0] {
		headers[i] = fmt.Sprintf("%v", header)
	}

	fieldMap := make(map[string]int)
	for i, header := range headers {
		fieldMap[header] = i
	}

	for rowIndex, row := range resp.Values[1:] {
		if !q.matchesWhere(row, headers, fieldMap) {
			continue
		}

		if q.offset > 0 && rowIndex < q.offset {
			continue
		}

		if q.limit > 0 && sliceValue.Len() >= q.limit {
			break
		}

		elem := reflect.New(elemType).Elem()
		if err := q.mapRowToStruct(row, headers, fieldMap, elem); err != nil {
			return fmt.Errorf("failed to map row to struct: %w", err)
		}

		sliceValue.Set(reflect.Append(sliceValue, elem))
	}

	return nil
}

func (q *Query) matchesWhere(row []interface{}, headers []string, fieldMap map[string]int) bool {
	for _, clause := range q.where {
		colIndex, exists := fieldMap[clause.Column]
		if !exists {
			continue
		}

		if colIndex >= len(row) {
			continue
		}

		cellValue := fmt.Sprintf("%v", row[colIndex])
		expectedValue := fmt.Sprintf("%v", clause.Value)

		switch clause.Operator {
		case "=", "==":
			if cellValue != expectedValue {
				return false
			}
		case "!=":
			if cellValue == expectedValue {
				return false
			}
		case ">":
			if !q.compareValues(cellValue, expectedValue, ">") {
				return false
			}
		case "<":
			if !q.compareValues(cellValue, expectedValue, "<") {
				return false
			}
		case ">=":
			if !q.compareValues(cellValue, expectedValue, ">=") {
				return false
			}
		case "<=":
			if !q.compareValues(cellValue, expectedValue, "<=") {
				return false
			}
		case "LIKE":
			if !strings.Contains(strings.ToLower(cellValue), strings.ToLower(expectedValue)) {
				return false
			}
		}
	}
	return true
}

func (q *Query) compareValues(a, b, operator string) bool {
	aFloat, aErr := strconv.ParseFloat(a, 64)
	bFloat, bErr := strconv.ParseFloat(b, 64)

	if aErr == nil && bErr == nil {
		switch operator {
		case ">":
			return aFloat > bFloat
		case "<":
			return aFloat < bFloat
		case ">=":
			return aFloat >= bFloat
		case "<=":
			return aFloat <= bFloat
		}
	}

	switch operator {
	case ">":
		return a > b
	case "<":
		return a < b
	case ">=":
		return a >= b
	case "<=":
		return a <= b
	}

	return false
}

func (q *Query) mapRowToStruct(row []interface{}, headers []string, fieldMap map[string]int, dest reflect.Value) error {
	destType := dest.Type()

	for i := 0; i < destType.NumField(); i++ {
		field := destType.Field(i)
		fieldValue := dest.Field(i)

		if !fieldValue.CanSet() {
			continue
		}

		tagValue := field.Tag.Get("sheet")
		if tagValue == "" {
			tagValue = field.Name
		}

		colIndex, exists := fieldMap[tagValue]
		if !exists || colIndex >= len(row) {
			continue
		}

		cellValue := fmt.Sprintf("%v", row[colIndex])
		if err := q.setFieldValue(fieldValue, cellValue); err != nil {
			return fmt.Errorf("failed to set field %s: %w", field.Name, err)
		}
	}

	return nil
}

func (q *Query) setFieldValue(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if value == "" {
			field.SetInt(0)
		} else {
			intVal, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return err
			}
			field.SetInt(intVal)
		}
	case reflect.Float32, reflect.Float64:
		if value == "" {
			field.SetFloat(0)
		} else {
			floatVal, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return err
			}
			field.SetFloat(floatVal)
		}
	case reflect.Bool:
		if value == "" {
			field.SetBool(false)
		} else {
			boolVal, err := strconv.ParseBool(value)
			if err != nil {
				return err
			}
			field.SetBool(boolVal)
		}
	default:
		return fmt.Errorf("unsupported field type: %s", field.Kind())
	}

	return nil
}

func (q *Query) Insert(data interface{}) error {
	dataValue := reflect.ValueOf(data)
	if dataValue.Kind() == reflect.Ptr {
		dataValue = dataValue.Elem()
	}

	if dataValue.Kind() != reflect.Struct {
		return fmt.Errorf("data must be a struct or pointer to struct")
	}

	readRange := fmt.Sprintf("%s!1:1", q.sheetName)
	resp, err := q.client.service.Spreadsheets.Values.Get(q.client.spreadsheetID, readRange).Do()
	if err != nil {
		return fmt.Errorf("failed to read headers: %w", err)
	}

	if len(resp.Values) == 0 {
		return fmt.Errorf("no headers found in sheet")
	}

	headers := make([]string, len(resp.Values[0]))
	for i, header := range resp.Values[0] {
		headers[i] = fmt.Sprintf("%v", header)
	}

	fieldMap := make(map[string]int)
	for i, header := range headers {
		fieldMap[header] = i
	}

	row := make([]interface{}, len(headers))
	dataType := dataValue.Type()

	for i := 0; i < dataType.NumField(); i++ {
		field := dataType.Field(i)
		fieldValue := dataValue.Field(i)

		tagValue := field.Tag.Get("sheet")
		if tagValue == "" {
			tagValue = field.Name
		}

		colIndex, exists := fieldMap[tagValue]
		if !exists {
			continue
		}

		row[colIndex] = fieldValue.Interface()
	}

	writeRange := fmt.Sprintf("%s!A:Z", q.sheetName)
	valueRange := &sheets.ValueRange{
		Values: [][]interface{}{row},
	}

	_, err = q.client.service.Spreadsheets.Values.Append(q.client.spreadsheetID, writeRange, valueRange).
		ValueInputOption("RAW").
		InsertDataOption("INSERT_ROWS").
		Do()

	if err != nil {
		return fmt.Errorf("failed to insert row: %w", err)
	}

	return nil
}

func (q *Query) Update(data interface{}) error {
	dataValue := reflect.ValueOf(data)
	if dataValue.Kind() == reflect.Ptr {
		dataValue = dataValue.Elem()
	}

	if dataValue.Kind() != reflect.Struct {
		return fmt.Errorf("data must be a struct or pointer to struct")
	}

	readRange := fmt.Sprintf("%s!A:Z", q.sheetName)
	resp, err := q.client.service.Spreadsheets.Values.Get(q.client.spreadsheetID, readRange).Do()
	if err != nil {
		return fmt.Errorf("failed to read sheet: %w", err)
	}

	if len(resp.Values) == 0 {
		return fmt.Errorf("no data found in sheet")
	}

	headers := make([]string, len(resp.Values[0]))
	for i, header := range resp.Values[0] {
		headers[i] = fmt.Sprintf("%v", header)
	}

	fieldMap := make(map[string]int)
	for i, header := range headers {
		fieldMap[header] = i
	}

	updatedRows := 0
	for rowIndex, row := range resp.Values[1:] {
		if !q.matchesWhere(row, headers, fieldMap) {
			continue
		}

		actualRowIndex := rowIndex + 2
		updatedRow := make([]interface{}, len(headers))
		copy(updatedRow, row)

		dataType := dataValue.Type()
		for i := 0; i < dataType.NumField(); i++ {
			field := dataType.Field(i)
			fieldValue := dataValue.Field(i)

			tagValue := field.Tag.Get("sheet")
			if tagValue == "" {
				tagValue = field.Name
			}

			colIndex, exists := fieldMap[tagValue]
			if !exists {
				continue
			}

			updatedRow[colIndex] = fieldValue.Interface()
		}

		updateRange := fmt.Sprintf("%s!A%d:Z%d", q.sheetName, actualRowIndex, actualRowIndex)
		valueRange := &sheets.ValueRange{
			Values: [][]interface{}{updatedRow},
		}

		_, err = q.client.service.Spreadsheets.Values.Update(q.client.spreadsheetID, updateRange, valueRange).
			ValueInputOption("RAW").
			Do()

		if err != nil {
			return fmt.Errorf("failed to update row %d: %w", actualRowIndex, err)
		}

		updatedRows++
	}

	if updatedRows == 0 {
		return fmt.Errorf("no rows matched the where conditions")
	}

	return nil
}

func (q *Query) Delete() error {
	readRange := fmt.Sprintf("%s!A:Z", q.sheetName)
	resp, err := q.client.service.Spreadsheets.Values.Get(q.client.spreadsheetID, readRange).Do()
	if err != nil {
		return fmt.Errorf("failed to read sheet: %w", err)
	}

	if len(resp.Values) == 0 {
		return fmt.Errorf("no data found in sheet")
	}

	headers := make([]string, len(resp.Values[0]))
	for i, header := range resp.Values[0] {
		headers[i] = fmt.Sprintf("%v", header)
	}

	fieldMap := make(map[string]int)
	for i, header := range headers {
		fieldMap[header] = i
	}

	var rowsToDelete []int
	for rowIndex, row := range resp.Values[1:] {
		if q.matchesWhere(row, headers, fieldMap) {
			actualRowIndex := rowIndex + 2
			rowsToDelete = append(rowsToDelete, actualRowIndex)
		}
	}

	if len(rowsToDelete) == 0 {
		return fmt.Errorf("no rows matched the where conditions")
	}

	for i := len(rowsToDelete) - 1; i >= 0; i-- {
		rowIndex := rowsToDelete[i]
		
		batchUpdateRequest := &sheets.BatchUpdateSpreadsheetRequest{
			Requests: []*sheets.Request{
				{
					DeleteDimension: &sheets.DeleteDimensionRequest{
						Range: &sheets.DimensionRange{
							SheetId:    q.getSheetId(),
							Dimension:  "ROWS",
							StartIndex: int64(rowIndex - 1),
							EndIndex:   int64(rowIndex),
						},
					},
				},
			},
		}

		_, err = q.client.service.Spreadsheets.BatchUpdate(q.client.spreadsheetID, batchUpdateRequest).Do()
		if err != nil {
			return fmt.Errorf("failed to delete row %d: %w", rowIndex, err)
		}
	}

	return nil
}

func (q *Query) getSheetId() int64 {
	resp, err := q.client.service.Spreadsheets.Get(q.client.spreadsheetID).Do()
	if err != nil {
		return 0
	}

	for _, sheet := range resp.Sheets {
		if sheet.Properties.Title == q.sheetName {
			return sheet.Properties.SheetId
		}
	}

	return 0
}