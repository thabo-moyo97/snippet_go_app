package services

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/jmoiron/sqlx"
	"thabomoyo.co.uk/internal/models"
)

type UserService struct {
	db     *sqlx.DB
	logger *slog.Logger
	model  *models.UserModel
}

func NewUserService(db *sqlx.DB, logger *slog.Logger) *UserService {
	return &UserService{
		db:     db,
		logger: logger,
		model:  &models.UserModel{DB: db},
	}
}

func (s *UserService) Authenticate(email, password string) (int, error) {
	return s.model.Authenticate(email, password)
}

func (s *UserService) Get(id int) (models.User, error) {
	return s.model.Get(id)
}

func (s *UserService) Insert(name, email, password string) error {
	return s.model.Insert(name, email, password)
}

func (s *UserService) Exists(id int) (bool, error) {
	return s.model.Exists(id)
}

func (s *UserService) IsAuthenticated(r *http.Request, sessions *SessionService) bool {
	return sessions.Get(r.Context(), "authenticatedUserID") != nil
}

func (s *UserService) GetAuthenticatedUser(r *http.Request, sessions *SessionService) (models.User, error) {
	userID := sessions.Get(r.Context(), "authenticatedUserID")
	if userID == nil {
		return models.User{}, errors.New("user not authenticated")
	}

	return s.Get(userID.(int))
}

func (s *UserService) Login(id int, sessions *SessionService, r *http.Request) error {
	sessions.Put(r.Context(), "authenticatedUserID", id)
	return nil
}

func (s *UserService) Logout(sessions *SessionService, r *http.Request) {
	sessions.Remove(r.Context(), "authenticatedUserID")
}
