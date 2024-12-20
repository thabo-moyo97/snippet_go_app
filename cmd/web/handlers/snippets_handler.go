package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"thabomoyo.co.uk/internal/models"
	"thabomoyo.co.uk/internal/services"
	"thabomoyo.co.uk/internal/validator"
)

type SnippetCreateViewForm struct {
	Title               string `form:"title"`
	Content             string `form:"content"`
	Expires             int    `form:"expires"`
	validator.Validator `form:"-"`
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
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := s.services.Templates.NewTemplateData(r)
	data.Snippets = snippets

	s.services.Templates.RenderView(w, r, http.StatusOK, "home", data)
}

func (s *SnippetHandler) SnippetView(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		s.services.Errors.ServerError(w, r, err)
		return
	}

	snippet, err := s.services.Snippets.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			s.services.Errors.ClientError(w, r, http.StatusNotFound)
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
	data.Form = SnippetCreateViewForm{
		Expires: 7,
	}

	s.services.Templates.RenderView(w, r, http.StatusOK, "snippets.create", data)
}

func (s *SnippetHandler) SnippetEditView(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		s.services.Errors.ServerError(w, r, err)
		return
	}
	data := s.services.Templates.NewTemplateData(r)
	snippet, err := s.services.Snippets.Get(id)
	data.Form = SnippetCreateViewForm{
		Title:   snippet.Title,
		Content: snippet.Content,
		Expires: 7,
	}

	s.services.Templates.RenderView(w, r, http.StatusOK, "snippets.edit", data)
}

func (s *SnippetHandler) SnippetCreatePostAction(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 4096)

	err := r.ParseForm()
	if err != nil {
		s.services.Errors.ClientError(w, r, http.StatusBadRequest)
		return
	}

	var form SnippetCreateViewForm

	err = s.services.Forms.DecodePostForm(r, &form)
	if err != nil {
		s.services.Errors.ClientError(w, r, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Title), "title", "This field cannot be blank")
	form.CheckField(validator.MaxChars(form.Title, 100), "title", "This field cannot be more than 100 characters long")
	form.CheckField(validator.NotBlank(form.Content), "content", "This field cannot be blank")
	form.CheckField(validator.MinChars(form.Content, 5), "content", "This field must be at least 5 characters long")
	form.CheckField(validator.MinWordCount(form.Content, 3), "content", "This field must contain at least 3 words")
	form.CheckField(validator.MaxChars(form.Content, 1000), "content", "This field must be less than 1000 characters long")
	form.CheckField(validator.PermittedValue(form.Expires, 1, 7, 365), "expires", "This field must equal 1, 7 or 365")

	if !form.Valid() {
		data := s.services.Templates.NewTemplateData(r)
		data.Form = form
		s.services.Templates.RenderView(w, r, http.StatusUnprocessableEntity, "snippets.create", data)
		return
	}

	id, err := s.services.Snippets.Insert(form.Title, form.Content, form.Expires)
	if err != nil {
		s.services.Logger.Error("failed to insert snippet", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	s.services.Sessions.Put(r.Context(), "flash", "Snippet successfully created!")

	http.Redirect(w, r, fmt.Sprintf("/snippet/view/%d", id), http.StatusSeeOther)
}
