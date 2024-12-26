package database

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

var ErrNoRecord = errors.New("record not found")

// DatabaseOperations provides common database operations
type DatabaseOperations struct {
	DB             *sqlx.DB
	TableName      string
	FillableFields []string
}

func NewDatabaseOperations(db *sqlx.DB, tableName string, fillableFields []string) *DatabaseOperations {
	return &DatabaseOperations{
		DB:             db,
		TableName:      tableName,
		FillableFields: fillableFields,
	}
}

func (ops *DatabaseOperations) Get(id int, dest interface{}) error {
	query := fmt.Sprintf("SELECT * FROM %s WHERE id = ?", ops.TableName)

	ops.DB.Select(dest, query)
	err := ops.DB.Get(dest, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNoRecord
		}
		return err
	}

	return nil
}

func (ops *DatabaseOperations) List(limit, offset int, dest interface{}) error {
	query := fmt.Sprintf("SELECT * FROM %s LIMIT ? OFFSET ?", ops.TableName)

	err := ops.DB.Select(&dest, query, limit, offset)
	if err != nil {
		return err
	}

	return nil
}

func (ops *DatabaseOperations) Exists(id int) (bool, error) {
	var exists bool
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE id = ?)", ops.TableName)

	err := ops.DB.QueryRow(query, id).Scan(&exists)
	return exists, err
}

func (ops *DatabaseOperations) Insert(data map[string]interface{}) (int, error) {
	query, args, err := squirrel.
		Insert(ops.TableName).
		SetMap(data).
		ToSql()

	if err != nil {
		return 0, fmt.Errorf("failed to build insert query: %w", err)
	}

	result, err := ops.DB.Exec(query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to execute insert query: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return int(id), nil
}

func (ops *DatabaseOperations) Update(id int, data map[string]interface{}) (bool, error) {
	tx, err := ops.DB.Begin()
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	query, args, err := squirrel.Update(ops.TableName).
		SetMap(data).
		Where("id = ?", id).
		ToSql()

	if err != nil {
		return false, err
	}

	_, err = tx.Exec(query, args...)
	if err != nil {
		return false, err
	}

	if err := tx.Commit(); err != nil {
		return false, err
	}

	return true, nil
}

func (ops *DatabaseOperations) Delete(id int) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE id = ?", ops.TableName)

	result, err := ops.DB.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrNoRecord
	}

	return nil
}

func (ops *DatabaseOperations) GetFillableFields() []string {
	return ops.FillableFields
}
