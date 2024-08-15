package handlers

import (
	"errors"
	"net/http"
	"thabomoyo.co.uk/cmd/web/config"
	"thabomoyo.co.uk/internal/models"
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
	App *config.Application
}

func (u *UserHandler) UserSignup(w http.ResponseWriter, r *http.Request) {
	data := u.App.NewTemplateData(r)
	data.Form = userSignupForm{}

	u.App.Render(w, r, http.StatusOK, "signup.tmpl", data)
}

func (u *UserHandler) UserLogin(w http.ResponseWriter, r *http.Request) {
	data := u.App.NewTemplateData(r)
	data.Form = userLoginForm{}
	u.App.Render(w, r, http.StatusOK, "login.tmpl", data)
}

func (u *UserHandler) UserSignupPost(w http.ResponseWriter, r *http.Request) {
	var form userSignupForm

	err := u.App.DecodePostForm(r, &form)
	if err != nil {
		u.App.ClientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Name), "name", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
	form.CheckField(validator.MinChars(form.Password, 8), "password", "This field must be at least 8 characters long")

	if !form.Valid() {
		data := u.App.NewTemplateData(r)
		data.Form = form
		u.App.Render(w, r, http.StatusUnprocessableEntity, "signup.tmpl", data)
		return
	}

	err = u.App.Users.Insert(form.Name, form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			form.AddFieldError("email", "Email address is already in use")

			data := u.App.NewTemplateData(r)
			data.Form = form
			u.App.Render(w, r, http.StatusUnprocessableEntity, "signup.tmpl", data)
		} else {
			u.App.ServerError(w, r, err)
		}

		return
	}

	u.App.SessionManager.Put(r.Context(), "flash", "Your signup was successful. Please log in.")

	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

func (u *UserHandler) UserLoginPost(w http.ResponseWriter, r *http.Request) {
	var form userLoginForm

	err := u.App.DecodePostForm(r, &form)
	if err != nil {
		u.App.ClientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")

	if !form.Valid() {
		data := u.App.NewTemplateData(r)
		data.Form = form
		u.App.Render(w, r, http.StatusUnprocessableEntity, "login.tmpl", data)
		return
	}

	id, err := u.App.Users.Authenticate(form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.AddNonFieldError("Email or password is incorrect")

			data := u.App.NewTemplateData(r)
			data.Form = form
			u.App.Render(w, r, http.StatusUnprocessableEntity, "login.tmpl", data)
		} else {
			u.App.ServerError(w, r, err)
		}
		return
	}

	err = u.App.SessionManager.RenewToken(r.Context())
	if err != nil {
		u.App.ServerError(w, r, err)
		return
	}

	u.App.SessionManager.Put(r.Context(), "authenticatedUserID", id)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (u *UserHandler) UserLogoutPost(w http.ResponseWriter, r *http.Request) {
	err := u.App.SessionManager.RenewToken(r.Context())
	if err != nil {
		u.App.ServerError(w, r, err)
		return
	}

	u.App.SessionManager.Remove(r.Context(), "authenticatedUserID")

	u.App.SessionManager.Put(r.Context(), "flash", "You've been logged out successfully!")

	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

func (u UserHandler) UserAccountView(w http.ResponseWriter, r *http.Request) {
	data := u.App.NewTemplateData(r)

	id := u.App.SessionManager.Get(r.Context(), "authenticatedUserID")

	user, err := u.App.Users.Get(id.(int))

	if err != nil {
		if errors.Is(models.ErrNoRecord, err) {
			u.App.SessionManager.Put(r.Context(), "flash", "User not found")
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
			return
		}

		u.App.ServerError(w, r, err)
		return
	}

	data.User = user

	u.App.Render(w, r, http.StatusOK, "account.tmpl", data)
}
