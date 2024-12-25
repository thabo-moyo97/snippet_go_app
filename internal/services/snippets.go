package services

import (
	"log/slog"

	"github.com/jmoiron/sqlx"
	"thabomoyo.co.uk/internal/models"
	"thabomoyo.co.uk/internal/utils"
)

type SnippetService struct {
	model  *models.SnippetModel
	logger *slog.Logger
}

func NewSnippetService(db *sqlx.DB, logger *slog.Logger) *SnippetService {
	return &SnippetService{
		model:  models.NewSnippetModel(db),
		logger: logger,
	}
}

func (s *SnippetService) Get(id int) (models.Snippet, error) {
	var snippet models.Snippet
	snippet, err := s.model.Get(id)
	return snippet, err
}

func (s *SnippetService) Create(snippet models.Snippet) (int, error) {
	fillable := s.model.GetFillableFields()
	data := utils.BuildDataMap(snippet, &fillable)
	return s.model.Insert(data)
}

// Insert is an alias for Create to maintain backwards compatibility
func (s *SnippetService) Insert(snippet models.Snippet) (int, error) {
	return s.Create(snippet)
}

func (s *SnippetService) Update(snippet models.Snippet) (bool, error) {
	fillable := s.model.GetFillableFields()
	data := utils.BuildDataMap(snippet, &fillable)
	return s.model.Update(snippet.ID, data)
}

func (s *SnippetService) Delete(id int) error {
	return s.model.Delete(id)
}

func (s *SnippetService) List(limit, offset int) ([]models.Snippet, error) {
	var snippets []models.Snippet
	snippets, err := s.model.List(limit, offset)
	return snippets, err
}

func (s *SnippetService) Exists(id int) (bool, error) {
	return s.model.Exists(id)
}

func (s *SnippetService) Latest() ([]models.Snippet, error) {
	return s.model.Latest()
}
