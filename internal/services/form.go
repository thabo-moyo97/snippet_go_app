package services

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-playground/form/v4"
)

type FormService struct {
	decoder *form.Decoder
	logger  *slog.Logger
}

func NewFormService(decoder *form.Decoder, logger *slog.Logger) *FormService {
	return &FormService{
		decoder: decoder,
		logger:  logger,
	}
}

func (s *FormService) DecodePostForm(r *http.Request, dst any) error {
	err := r.ParseForm()
	if err != nil {
		s.logger.Error("failed to parse form", "error", err)
		return err
	}

	err = s.decoder.Decode(dst, r.PostForm)
	if err != nil {
		var invalidDecoderError *form.InvalidDecoderError
		if errors.As(err, &invalidDecoderError) {
			s.logger.Error("invalid decoder error", "error", err)
			panic(err)
		}
		s.logger.Error("failed to decode form", "error", err)
		return err
	}

	return nil
}

// Helper method to validate form fields
func (s *FormService) ValidateField(field string, value string, validations ...func(string) bool) []string {
	var errors []string
	for _, validation := range validations {
		if !validation(value) {
			errors = append(errors, field+" is invalid")
		}
	}
	return errors
}

// Helper method to check if form has any errors
func (s *FormService) HasErrors(errors map[string][]string) bool {
	return len(errors) > 0
}
