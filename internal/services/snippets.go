package services

import (
	"database/sql"
	"log/slog"
	"time"

	"thabomoyo.co.uk/internal/models"
)

type SnippetService struct {
	db     *sql.DB
	logger *slog.Logger
	model  *models.SnippetModel
}

func NewSnippetService(db *sql.DB, logger *slog.Logger) *SnippetService {
	return &SnippetService{
		db:     db,
		logger: logger,
		model:  &models.SnippetModel{DB: db},
	}
}

// Move business logic from handlers to here
func (s *SnippetService) Latest() ([]models.Snippet, error) {
	return s.model.Latest()
}

func (s *SnippetService) Get(id int) (models.Snippet, error) {
	return s.model.Get(id)
}

func (s *SnippetService) Insert(title, content string, expires int) (int, error) {
	return s.model.Insert(title, content, expires)
}
func (s *SnippetService) Update(id int, title string, content string, expires int) (int, error) {
	snippet := models.Snippet{
		ID:      id,
		Title:   title,
		Content: content,
		Expires: time.Now().AddDate(0, 0, expires),
	}

	return s.model.Update(snippet)
}
