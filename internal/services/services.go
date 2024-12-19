package services

import (
	"database/sql"
	"html/template"
	"log/slog"
	"os"

	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	"thabomoyo.co.uk/internal/templatemanager"
)

// Services holds all services used by the application
type Services struct {
	Snippets      *SnippetService
	Users         *UserService
	Templates     *TemplateService
	Sessions      *scs.SessionManager
	Forms         *FormService
	Errors        *ErrorService
	Logger        *slog.Logger
	DebugMode     bool
	IsDevelopment bool
	Watcher       *Watcher
}

// NewServices creates and initialises all services
func NewServices(
	db *sql.DB,
	logger *slog.Logger,
	templateCache map[string]*template.Template,
	sessionManager *scs.SessionManager,
	templateManager *templatemanager.Manager,
	debugMode bool,
) *Services {

	isDevelopment := os.Getenv("APP_MODE") == "local"
	templateService := NewTemplateService(templateCache, templateManager, sessionManager, logger, isDevelopment)
	formDecoder := form.NewDecoder()

	watcher, err := NewWatcher("./ui/html", templateManager, templateService)
	if err != nil {
		logger.Error("failed to initialise watcher", "error", err)
	}

	return &Services{
		Snippets:      NewSnippetService(db, logger),
		Users:         NewUserService(db, logger),
		Templates:     templateService,
		Sessions:      sessionManager,
		Forms:         NewFormService(formDecoder, logger),
		Errors:        NewErrorService(logger, templateService, debugMode),
		Logger:        logger,
		DebugMode:     debugMode,
		IsDevelopment: isDevelopment,
		Watcher:       watcher,
	}
}
