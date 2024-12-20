package handlers

import (
	"errors"
	"net/http"

	"thabomoyo.co.uk/internal/models"
	"thabomoyo.co.uk/internal/services"
	"thabomoyo.co.uk/internal/validator"
)

type userSignupForm struct {
	Name                string `form:"name"`
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

type userLoginForm struct {
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

type UserHandler struct {
	services *services.Services
}

func NewUserHandler(services *services.Services) *UserHandler {
	return &UserHandler{services: services}
}

func (u *UserHandler) UserSignupView(w http.ResponseWriter, r *http.Request) {
	data := u.services.Templates.NewTemplateData(r)
	data.Form = userSignupForm{}

	u.services.Templates.RenderView(w, r, http.StatusOK, "signup", data)
}

func (u *UserHandler) UserLoginView(w http.ResponseWriter, r *http.Request) {
	data := u.services.Templates.NewTemplateData(r)
	data.Form = userLoginForm{}

	u.services.Templates.RenderView(w, r, http.StatusOK, "login", data)
}

func (u *UserHandler) UserSignupPostAction(w http.ResponseWriter, r *http.Request) {
	var form userSignupForm

	err := u.services.Forms.DecodePostForm(r, &form)
	if err != nil {
		u.services.Errors.ServerError(w, r, err)
		return
	}

	form.CheckField(validator.NotBlank(form.Name), "name", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
	form.CheckField(validator.MinChars(form.Password, 8), "password", "This field must be at least 8 characters long")

	if !form.Valid() {
		data := u.services.Templates.NewTemplateData(r)
		data.Form = form
		u.services.Templates.RenderView(w, r, http.StatusUnprocessableEntity, "signup", data)
		return
	}

	err = u.services.Users.Insert(form.Name, form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) || errors.Is(err, models.ErrUserExists) {
			form.AddFieldError("email", "Email address is already in use")

			data := u.services.Templates.NewTemplateData(r)
			data.Form = form
			u.services.Templates.RenderView(w, r, http.StatusUnprocessableEntity, "signup", data)
		} else {
			u.services.Errors.ServerError(w, r, err)
		}

		return
	}

	u.services.Sessions.Put(r.Context(), "flash", "Your signup was successful. Please log in.")

	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

func (u *UserHandler) UserLoginPostAction(w http.ResponseWriter, r *http.Request) {
	var form userLoginForm

	err := u.services.Forms.DecodePostForm(r, &form)
	if err != nil {
		u.services.Errors.ClientError(w, r, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")

	if !form.Valid() {
		data := u.services.Templates.NewTemplateData(r)
		data.Form = form
		u.services.Templates.RenderView(w, r, http.StatusUnprocessableEntity, "login", data)
		return
	}

	id, err := u.services.Users.Authenticate(form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.AddNonFieldError("Email or password is incorrect")

			data := u.services.Templates.NewTemplateData(r)
			data.Form = form
			u.services.Templates.RenderView(w, r, http.StatusUnprocessableEntity, "login", data)
		} else {
			u.services.Errors.ServerError(w, r, err)
		}
		return
	}

	err = u.services.Sessions.RenewToken(r.Context())
	if err != nil {
		u.services.Errors.ServerError(w, r, err)
		return
	}

	u.services.Sessions.Put(r.Context(), "authenticatedUserID", id)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (u *UserHandler) UserLogoutPostAction(w http.ResponseWriter, r *http.Request) {
	err := u.services.Sessions.RenewToken(r.Context())
	if err != nil {
		u.services.Errors.ServerError(w, r, err)
		return
	}

	u.services.Sessions.Remove(r.Context(), "authenticatedUserID")

	u.services.Sessions.Put(r.Context(), "flash", "You've been logged out successfully!")

	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

func (u *UserHandler) UserAccountView(w http.ResponseWriter, r *http.Request) {
	data := u.services.Templates.NewTemplateData(r)

	id := u.services.Sessions.Get(r.Context(), "authenticatedUserID")

	user, err := u.services.Users.Get(id.(int))

	if err != nil {
		if errors.Is(models.ErrNoRecord, err) {
			u.services.Sessions.Put(r.Context(), "flash", "User not found")
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
			return
		}

		u.services.Errors.ServerError(w, r, err)
		return
	}

	data.User = user
	u.services.Templates.RenderView(w, r, http.StatusOK, "account", data)
}
