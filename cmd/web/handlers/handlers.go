package handlers

import "thabomoyo.co.uk/internal/services"

// Handlers holds all HTTP handlers
type Handlers struct {
	snippets *SnippetHandler
	users    *UserHandler
}

// NewHandlers creates a new Handlers instance
func NewHandlers(services *services.Services) *Handlers {
	return &Handlers{
		snippets: NewSnippetHandler(services),
		users:    NewUserHandler(services),
	}
}

func (h *Handlers) Snippets() *SnippetHandler {
	return h.snippets
}

func (h *Handlers) Users() *UserHandler {
	return h.users
}
