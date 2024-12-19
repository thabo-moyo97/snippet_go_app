package models

import "errors"

var (
	ErrNoRecord           = errors.New("models: no matching record found")
	ErrInvalidCredentials = errors.New("models: invalid credentials")
	ErrDuplicateEmail     = errors.New("models: duplicate email")
	ErrUserExists         = errors.New("models: user already exists")
	ErrTableNotExists     = errors.New("models: table does not exist")
	ErrUserNotFound       = errors.New("models: user not found")
)
