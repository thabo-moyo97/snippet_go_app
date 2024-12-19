package services

import (
	"context"

	"github.com/alexedwards/scs/v2"
)

type SessionService struct {
	manager *scs.SessionManager
}

func NewSessionService(manager *scs.SessionManager) *SessionService {
	return &SessionService{
		manager: manager,
	}
}

func (s *SessionService) Put(ctx context.Context, key string, val interface{}) {
	s.manager.Put(ctx, key, val)
}

func (s *SessionService) Get(ctx context.Context, key string) interface{} {
	return s.manager.Get(ctx, key)
}

func (s *SessionService) Remove(ctx context.Context, key string) {
	s.manager.Remove(ctx, key)
}

func (s *SessionService) Exists(ctx context.Context, key string) bool {
	return s.manager.Exists(ctx, key)
}

func (s *SessionService) RenewToken(ctx context.Context) error {
	return s.manager.RenewToken(ctx)
}
