package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"thabomoyo.co.uk/internal/utils"

	"thabomoyo.co.uk/internal/models"
	"thabomoyo.co.uk/internal/services"
	"thabomoyo.co.uk/internal/validator"
)

type SnippetCreateViewForm struct {
	Title     string `form:"title" validate:"required,min=3"`
	Content   string `form:"content" validate:"required,min=10"`
	ExpiresAt int    `form:"expires_at" validate:"oneof=0 1 7 28 365 never"`
	*validator.Form
}

type SnippetViewForm struct {
	ID        int          `form:"id" validate:"exists=snippets:id"`
	Title     string       `form:"title" validate:"required,min=3"`
	Content   string       `form:"content" validate:"required,min=10"`
	ExpiresAt sql.NullTime `form:"expires_at"`
	*validator.Form
}

type SnippetUpdateViewForm struct {
	ID        int    `form:"id" validate:"required,min=1,exists=snippets:id"`
	Title     string `form:"title" validate:"required,min=3"`
	Content   string `form:"content" validate:"required,min=10"`
	ExpiresAt string `form:"expires_at"`
	*validator.Form
}

type SnippetDeleteForm struct {
	ID int `form:"id" validate:"required,min=1,exists=snippets:id"`
	*validator.Form
}

type SnippetHandler struct {
	services *services.Services
}

func NewSnippetHandler(services *services.Services) *SnippetHandler {
	return &SnippetHandler{
		services: services,
	}
}

func (s *SnippetHandler) Home(w http.ResponseWriter, r *http.Request) {
	snippets, err := s.services.Snippets.Latest()
	if err != nil {
		s.services.Logger.Error("failed to get latest snippets", "error", err)
		s.services.Errors.ServerError(w, r, err, http.StatusInternalServerError)
		return
	}

	data := s.services.Templates.NewTemplateData(r)
	data.Snippets = snippets

	s.services.Templates.RenderView(w, r, http.StatusOK, "home", data)
}

func (s *SnippetHandler) SnippetShow(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 0 {
		s.services.Errors.ServerError(w, r, err, http.StatusNotFound)
		return
	}

	snippet, err := s.services.Snippets.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			s.services.Errors.ClientError(w, r, http.StatusNotFound, &err)
			return
		} else {
			s.services.Logger.Error("failed to get snippet", "error", err)
			s.services.Errors.ServerError(w, r, err)
		}
		return
	}

	data := s.services.Templates.NewTemplateData(r)
	data.Snippet = snippet

	s.services.Templates.RenderView(w, r, http.StatusOK, "snippets.view", data)
}

func (s *SnippetHandler) SnippetCreateView(w http.ResponseWriter, r *http.Request) {
	data := s.services.Templates.NewTemplateData(r)
	data.Form = &SnippetCreateViewForm{
		Form:      &validator.Form{},
		ExpiresAt: 0,
	}

	s.services.Templates.RenderView(w, r, http.StatusOK, "snippets.create", data)
}

func (s *SnippetHandler) SnippetEditView(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		s.services.Errors.ServerError(w, r, err)
		return
	}
	data := s.services.Templates.NewTemplateData(r)
	snippet, err := s.services.Snippets.Get(id)
	if err != nil {
		s.services.Errors.ServerError(w, r, err)
		return
	}
	data.Form = &SnippetViewForm{
		ID:        snippet.ID,
		Title:     snippet.Title,
		Content:   snippet.Content,
		ExpiresAt: snippet.ExpiresAt,
		Form:      &validator.Form{},
	}

	s.services.Templates.RenderView(w, r, http.StatusOK, "snippets.edit", data)
}

func (s *SnippetHandler) SnippetCreateAction(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 4096)

	var form SnippetCreateViewForm

	err := s.services.Forms.DecodePostForm(r, &form)

	if err != nil {
		s.services.Errors.ClientError(w, r, http.StatusBadRequest, &err)
		return
	}

	form.Form = s.services.Validator.Validate(&form)

	if !form.Valid() {
		data := s.services.Templates.NewTemplateData(r)
		data.Form = form
		s.services.Templates.RenderView(w, r, http.StatusUnprocessableEntity, "snippets.create", data)
		return
	}

	// Form is valid, create the snippet
	snippet := models.Snippet{
		Title:   form.Title,
		Content: form.Content,
		ExpiresAt: sql.NullTime{
			Time:  time.Now().AddDate(0, 0, form.ExpiresAt),
			Valid: form.ExpiresAt > 0,
		},
	}

	err = utils.BindFormData(&form, &snippet)

	id, insertErr := s.services.Snippets.Insert(snippet)

	if insertErr != nil {
		s.services.Logger.Error("failed to insert snippet", "error", insertErr)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	s.services.Sessions.Put(r.Context(), "flash", "Snippet successfully created!")

	http.Redirect(w, r, fmt.Sprintf("/snippet/%d/view", id), http.StatusSeeOther)
}

func (s *SnippetHandler) SnippetEditAction(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	id, _ := strconv.Atoi(r.PathValue("id"))

	var form SnippetUpdateViewForm
	form.ID = id

	err := s.services.Forms.DecodePostForm(r, &form)
	if err != nil {
		s.services.Errors.ClientError(w, r, http.StatusBadRequest, &err)
		return
	}

	form.Form = s.services.Validator.Validate(&form)

	if !form.Valid() {
		data := s.services.Templates.NewTemplateData(r)
		data.Form = form
		s.services.Templates.RenderView(w, r, http.StatusUnprocessableEntity, "snippets.edit", data)
		return
	}

	expiry, _ := time.Parse("02 Jan 2006 at 15:04", form.ExpiresAt)

	snippet := models.Snippet{
		ExpiresAt: sql.NullTime{
			Time:  expiry,
			Valid: true,
		},
	}

	utils.BindFormData(&form, &snippet)

	success, err := s.services.Snippets.Update(snippet)

	if err != nil || !success {
		s.services.Logger.Error("failed to insert snippet", "error", err)
		s.services.Errors.ServerError(w, r, err)
		return
	}

	s.services.Sessions.Put(r.Context(), "flash", "Snippet successfully updated!")

	http.Redirect(w, r, fmt.Sprintf("/snippet/%d/view", id), http.StatusSeeOther)

}

func (s *SnippetHandler) SnippetDeleteAction(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		s.services.Errors.ClientError(w, r, http.StatusBadRequest, &err)
		return
	}

	// Validate the snippet exists
	form := SnippetDeleteForm{
		ID:   id,
		Form: &validator.Form{},
	}

	form.Form = s.services.Validator.Validate(&form)
	if !form.Valid() {
		s.services.Errors.ClientError(w, r, http.StatusNotFound, nil)
		return
	}

	// Delete the snippet
	err = s.services.Snippets.Delete(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			s.services.Errors.ClientError(w, r, http.StatusNotFound, &err)
			return
		}
		s.services.Logger.Error("failed to delete snippet", "error", err)
		s.services.Errors.ServerError(w, r, err)
		return
	}

	s.services.Sessions.Put(r.Context(), "flash", "Snippet successfully deleted!")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
