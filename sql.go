package sheetsql

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type SQLParser struct {
	client *Client
}

func NewSQLParser(client *Client) *SQLParser {
	return &SQLParser{client: client}
}

func (p *SQLParser) Query(sql string, dest interface{}) error {
	query, err := p.parseSQL(sql)
	if err != nil {
		return fmt.Errorf("failed to parse SQL: %w", err)
	}

	return query.Get(dest)
}

func (p *SQLParser) parseSQL(sql string) (*Query, error) {
	sql = strings.TrimSpace(sql)
	sql = regexp.MustCompile(`\s+`).ReplaceAllString(sql, " ")

	selectRegex := regexp.MustCompile(`(?i)^SELECT\s+(.+?)\s+FROM\s+(\w+)(?:\s+WHERE\s+(.+?))?(?:\s+LIMIT\s+(\d+))?(?:\s+OFFSET\s+(\d+))?$`)
	matches := selectRegex.FindStringSubmatch(sql)

	if len(matches) == 0 {
		return nil, fmt.Errorf("invalid SQL syntax")
	}

	tableName := matches[2]
	query := p.client.From(tableName)

	if matches[3] != "" {
		whereClause := matches[3]
		if err := p.parseWhere(query, whereClause); err != nil {
			return nil, fmt.Errorf("failed to parse WHERE clause: %w", err)
		}
	}

	if matches[4] != "" {
		limit, err := strconv.Atoi(matches[4])
		if err != nil {
			return nil, fmt.Errorf("invalid LIMIT value: %w", err)
		}
		query.Limit(limit)
	}

	if matches[5] != "" {
		offset, err := strconv.Atoi(matches[5])
		if err != nil {
			return nil, fmt.Errorf("invalid OFFSET value: %w", err)
		}
		query.Offset(offset)
	}

	return query, nil
}

func (p *SQLParser) parseWhere(query *Query, whereClause string) error {
	conditions := regexp.MustCompile(`(?i)\s+AND\s+`).Split(whereClause, -1)

	for _, condition := range conditions {
		condition = strings.TrimSpace(condition)

		operatorRegex := regexp.MustCompile(`(\w+)\s*(=|!=|<>|<=|>=|<|>|LIKE)\s*(.+)`)
		matches := operatorRegex.FindStringSubmatch(condition)

		if len(matches) != 4 {
			return fmt.Errorf("invalid WHERE condition: %s", condition)
		}

		column := matches[1]
		operator := matches[2]
		value := strings.Trim(matches[3], "'\"")

		if operator == "<>" {
			operator = "!="
		}

		var parsedValue interface{}
		if intVal, err := strconv.Atoi(value); err == nil {
			parsedValue = intVal
		} else if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			parsedValue = floatVal
		} else if boolVal, err := strconv.ParseBool(value); err == nil {
			parsedValue = boolVal
		} else {
			parsedValue = value
		}

		query.Where(column, operator, parsedValue)
	}

	return nil
}

func (p *SQLParser) Insert(sql string, data interface{}) error {
	return fmt.Errorf("SQL INSERT not implemented yet")
}

func (p *SQLParser) Update(sql string, data interface{}) error {
	return fmt.Errorf("SQL UPDATE not implemented yet")
}

func (p *SQLParser) Delete(sql string) error {
	return fmt.Errorf("SQL DELETE not implemented yet")
}
