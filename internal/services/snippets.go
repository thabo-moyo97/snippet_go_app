package services

import (
	"log/slog"

	"github.com/jmoiron/sqlx"
	"thabomoyo.co.uk/internal/models"
	"thabomoyo.co.uk/internal/utils"
)

type SnippetService struct {
	snippet models.SnippetModelInterface
	logger  *slog.Logger
}

func NewSnippetService(db *sqlx.DB, logger *slog.Logger) *SnippetService {
	return &SnippetService{
		snippet: models.NewSnippetModel(db),
		logger:  logger,
	}
}

func (s *SnippetService) Get(id int) (models.Snippet, error) {
	model, err := s.snippet.Get(id)

	return model, err
}

func (s *SnippetService) Create(snippet models.Snippet) (int, error) {
	return s.snippet.Insert(snippet)
}

func (s *SnippetService) Insert(snippet models.Snippet) (int, error) {

	return s.snippet.Insert(snippet)
}

func (s *SnippetService) Update(snippet models.Snippet) (bool, error) {
	fillable := s.snippet.GetFillableFields()
	data := utils.BuildDataMap(snippet, &fillable)

	return s.snippet.Update(snippet.ID, data)
}

func (s *SnippetService) Delete(id int) error {
	return s.snippet.Delete(id)
}

func (s *SnippetService) List(limit, offset int) ([]models.Snippet, error) {
	return s.snippet.List(limit, offset)
}

func (s *SnippetService) Exists(id int) (bool, error) {
	return s.snippet.Exists(id)
}

func (s *SnippetService) Latest() ([]models.Snippet, error) {
	return s.snippet.Latest()
}
