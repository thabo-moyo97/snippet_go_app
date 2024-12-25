package models

import (
	"database/sql"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/jmoiron/sqlx"
)

type Snippet struct {
	ID        int
	Title     string
	Content   string
	ExpiresAt sql.NullTime `db:"expires_at"`
	CreatedAt time.Time    `db:"created_at"`
}

type SnippetModelInterface interface {
	ModelOperations[Snippet]
	Latest() ([]Snippet, error)
}
type SnippetModel struct {
	*Model[Snippet]
}

func NewSnippetModel(db *sqlx.DB) *SnippetModel {
	table := "snippets"
	fillableFields := []string{"title", "content", "expires_at"}
	return &SnippetModel{
		Model: NewModel[Snippet](db, table, fillableFields),
	}
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

	err = m.DB.Select(&snippets, query, args...)
	if err != nil {
		return nil, err
	}

	return snippets, nil
}
