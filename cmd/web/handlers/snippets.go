package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"thabomoyo.co.uk/cmd/web/config"
	"thabomoyo.co.uk/internal/models"
	"thabomoyo.co.uk/internal/validator"
)

type snippetCreateForm struct {
	Title               string `form:"title"`
	Content             string `form:"content"`
	Expires             int    `form:"expires"`
	validator.Validator `form:"-"`
}

type SnippetHandler struct {
	App *config.Application
}

func (s *SnippetHandler) Home(w http.ResponseWriter, r *http.Request) {
	snippets, err := s.App.Snippets.Latest()

	if err != nil {
		s.App.ServerError(w, r, err)
		return
	}
	data := s.App.NewTemplateData(r)
	data.Snippets = snippets

	s.App.Render(w, r, http.StatusOK, "home.tmpl", data)
}

func (s *SnippetHandler) SnippetView(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	snippet, err := s.App.Snippets.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(w, r)
		} else {
			s.App.ServerError(w, r, err)
		}
		return
	}

	data := s.App.NewTemplateData(r)
	data.Snippet = snippet

	s.App.Render(w, r, http.StatusOK, "view.tmpl", data)
}

func (s *SnippetHandler) SnippetCreate(w http.ResponseWriter, r *http.Request) {
	data := s.App.NewTemplateData(r)
	data.Form = snippetCreateForm{
		Expires: 7,
	}

	s.App.Render(w, r, http.StatusOK, "create.tmpl", data)
}

func (s *SnippetHandler) SnippetCreatePost(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 4096)

	err := r.ParseForm()
	if err != nil {
		s.App.ClientError(w, http.StatusBadRequest)
		return
	}

	var form snippetCreateForm

	err = s.App.FormDecoder.Decode(&form, r.PostForm)
	if err != nil {
		s.App.ClientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Title), "title", "This field cannot be blank")
	form.CheckField(validator.MaxChars(form.Title, 100), "title", "This field cannot be more than 100 characters long")
	form.CheckField(validator.NotBlank(form.Content), "content", "This field cannot be blank")
	form.CheckField(validator.MinChars(form.Content, 10), "content", "This field must be at least 10 characters long")
	form.CheckField(validator.MinWordCount(form.Content, 5), "content", "This field must contain at least 5 words")
	form.CheckField(validator.MaxChars(form.Content, 1000), "content", "This field must be less than 1000 characters long")
	form.CheckField(validator.PermittedValue(form.Expires, 1, 7, 365), "expires", "This field must equal 1, 7 or 365")

	if !form.Valid() {
		data := s.App.NewTemplateData(r)
		data.Form = form
		s.App.Render(w, r, http.StatusUnprocessableEntity, "create.tmpl", data)
		return
	}

	id, err := s.App.Snippets.Insert(form.Title, form.Content, form.Expires)
	if err != nil {
		s.App.ServerError(w, r, err)
		return
	}

	s.App.SessionManager.Put(r.Context(), "flash", "Snippet successfully created!")

	http.Redirect(w, r, fmt.Sprintf("/snippet/view/%d", id), http.StatusSeeOther)
}
