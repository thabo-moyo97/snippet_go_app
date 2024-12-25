package models

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"

	"thabomoyo.co.uk/internal/utils"
)

// ModelData represents any database entity
type ModelData interface{}

// Reader defines read-only database operations
type Reader[T ModelData] interface {
	Get(id int) (T, error)
	List(limit, offset int) ([]T, error)
	Exists(id int) (bool, error)
	GetFillableFields() []string
}

// Writer defines write database operations
type Writer[T ModelData] interface {
	Insert(data T) (int, error)
	Update(id int, data map[string]interface{}) (bool, error)
	Delete(id int) error
}

// Repository combines read and write operations
type ModelOperations[T ModelData] interface {
	Reader[T]
	Writer[T]
}

// Implementation types
type Model[T ModelData] struct {
	DB             *sqlx.DB
	TableName      string
	FillableFields []string
}

func NewModel[T ModelData](db *sqlx.DB, tableName string, fillableFields []string) *Model[T] {
	return &Model[T]{
		DB:             db,
		TableName:      tableName,
		FillableFields: fillableFields,
	}
}

// Reader implementations
func (m *Model[T]) Get(id int) (T, error) {
	var result T
	query := fmt.Sprintf("SELECT * FROM %s WHERE id = ?", m.TableName)

	err := m.DB.Get(&result, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return result, ErrNoRecord
		}
		return result, err
	}

	return result, nil
}

func (m *Model[T]) List(limit, offset int) ([]T, error) {
	var results []T
	query := fmt.Sprintf("SELECT * FROM %s LIMIT ? OFFSET ?", m.TableName)

	err := m.DB.Select(&results, query, limit, offset)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func (m *Model[T]) Exists(id int) (bool, error) {
	var exists bool
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE id = ?)", m.TableName)

	err := m.DB.QueryRow(query, id).Scan(&exists)
	return exists, err
}

// Writer implementations
func (m *Model[T]) Insert(model T) (int, error) {
	data := utils.BuildDataMap(model, &m.FillableFields)

	query, args, err := squirrel.
		Insert(m.TableName).
		SetMap(data).
		ToSql()

	if err != nil {
		return 0, fmt.Errorf("failed to build insert query: %w", err)
	}

	result, err := m.DB.Exec(query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to execute insert query: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return int(id), nil
}

func (m *Model[T]) Update(id int, data map[string]interface{}) (bool, error) {

	tx, err := m.DB.Begin()

	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	query, args, err := squirrel.Update(m.TableName).
		SetMap(data).
		Where("id = ?", id).
		ToSql()

	if err != nil {
		return false, err
	}

	_, err = tx.Exec(query, args...)

	if err := tx.Commit(); err != nil {
		return false, err
	}

	return true, nil
}

func (m *Model[T]) Delete(id int) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE id = ?", m.TableName)

	result, err := m.DB.Exec(query, id)
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

func (m *Model[T]) GetFillableFields() []string {
	return m.FillableFields
}
