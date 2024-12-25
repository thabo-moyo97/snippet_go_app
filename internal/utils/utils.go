package utils

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"
	"unicode"
)

// BuildInsertQuery dynamically constructs an INSERT SQL query
func BuildInsertQuery(table string, data map[string]interface{}) (string, []interface{}) {
	var columns []string
	var placeholders []string
	var args []interface{}

	for col, val := range data {
		columns = append(columns, col)
		if val == nil {
			placeholders = append(placeholders, "NULL")
		} else {
			placeholders = append(placeholders, "?")
			args = append(args, formatValue(val))
		}
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, strings.Join(columns, ", "), strings.Join(placeholders, ", "))
	return query, args
}

// BuildUpdateQuery dynamically constructs an UPDATE SQL query
func BuildUpdateQuery(table string, data map[string]interface{}, condition string, conditionArgs ...interface{}) (string, []interface{}) {
	var setClauses []string
	var args []interface{}

	for col, val := range data {
		if strings.Compare(strings.ToLower(col), strings.ToLower(condition)) == 0 {
			continue
		}

		if val != nil {
			setClauses = append(setClauses, fmt.Sprintf("%s = ?", col))
			args = append(args, formatValue(val))
		}
	}

	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s", table, strings.Join(setClauses, ", "), condition)
	args = append(args, conditionArgs...)
	return query, args
}

// formatValue formats special values like dates and datetimes
func formatValue(value interface{}) interface{} {
	switch v := value.(type) {
	case time.Time:
		return v.Format("2006-01-02 15:04:05")
	default:
		return value
	}
}

func BuildDataMap(model interface{}, fillable *[]string) map[string]interface{} {
	data := make(map[string]interface{})

	val := reflect.ValueOf(model)
	typ := reflect.TypeOf(model)

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldName := strings.ToLower(toSnakeCase(typ.Field(i).Name))

		if fillable == nil || contains(*fillable, fieldName) {
			data[fieldName] = field.Interface()
		}
	}

	return data
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(unicode.ToLower(r))
	}
	return result.String()
}

// BindFormData binds form data to a struct
func BindFormData(form interface{}, target interface{}) error {
	var bindErr error
	defer func() {
		if err := recover(); err != nil {
			bindErr = fmt.Errorf("binding error: %v", err)
		}
	}()
	formVal := reflect.ValueOf(form).Elem()
	targetVal := reflect.ValueOf(target).Elem()

	for i := 0; i < formVal.NumField(); i++ {
		formField := formVal.Field(i)
		targetField := targetVal.FieldByName(formVal.Type().Field(i).Name)

		if targetField.IsValid() && targetField.CanSet() && targetField.Type() == formField.Type() {
			targetField.Set(formField)
		}
	}

	return bindErr
}

func getPrimaryKeyColumnName(db *sql.DB, tableName string) (string, error) {
	query := `
        SELECT COLUMN_NAME
        FROM INFORMATION_SCHEMA.COLUMNS
        WHERE TABLE_NAME = ? AND COLUMN_KEY = 'PRI'
    `
	var primaryKey string
	err := db.QueryRow(query, tableName).Scan(&primaryKey)
	if err != nil {
		return "", fmt.Errorf("could not get primary key for table %s: %v", tableName, err)
	}
	return primaryKey, nil
}
