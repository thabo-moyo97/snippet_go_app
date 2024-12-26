package models

import (
	"database/sql"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"thabomoyo.co.uk/internal/database"
)

type Snippet struct {
	ID        int
	Title     string
	Content   string
	ExpiresAt sql.NullTime `db:"expires_at"`
	CreatedAt time.Time    `db:"created_at"`
}

// SnippetModelInterface defines the contract for snippet operations
type SnippetModelInterface interface {
	database.Operations
	Get(id int) (Snippet, error)               // Override with specific return type
	List(limit, offset int) ([]Snippet, error) // Override with specific return type
	Latest() ([]Snippet, error)                // Snippet-specific method
}

type SnippetModel struct {
	*database.DatabaseOperations
}

var dbTable = "snippets"
var fillable = []string{"title", "content", "expires_at"}

func NewSnippetModel(db *sqlx.DB) SnippetModelInterface {
	return &SnippetModel{
		DatabaseOperations: database.NewDatabaseOperations(
			db,
			dbTable,
			fillable,
		),
	}
}

// Get retrieves a specific snippet
func (m *SnippetModel) Get(id int) (Snippet, error) {
	var snippet Snippet
	err := m.DatabaseOperations.GetByID(id, &snippet)
	return snippet, err
}

// List retrieves a list of snippets
func (m *SnippetModel) List(limit, offset int) ([]Snippet, error) {
	var snippets []Snippet
	err := m.DatabaseOperations.ListPagination(limit, offset, &snippets)
	return snippets, err
}

// Latest is a custom method specific to Snippet
func (m *SnippetModel) Latest() ([]Snippet, error) {
	var snippets []Snippet
	query, args, err := sq.
		Select("*").
		From(m.TableName).
		OrderBy("id DESC").
		Limit(10).
		ToSql()

	if err != nil {
		return nil, err
	}

	err = m.DB.Select(&snippets, query, args...)
	if err != nil {
		return nil, err
	}

	return snippets, nil
}
