package services

import (
	"log/slog"
	"os"
	"time"

	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	"github.com/jmoiron/sqlx"
	"thabomoyo.co.uk/internal/templatemanager"
	"thabomoyo.co.uk/internal/validator"
)

// Services holds all services used by the application
type Services struct {
	DB            *sqlx.DB
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
	Validator     *validator.Validator
}

// NewServices creates and initialises all services
func NewServices(
	db *sqlx.DB,
	logger *slog.Logger,
	debugMode bool,
) *Services {

	isDevelopment := os.Getenv("APP_MODE") == "local" || debugMode
	sessionManager := scs.New()
	sessionManager.Store = mysqlstore.New(db.DB)
	sessionManager.Lifetime = 12 * time.Hour
	sessionManager.Cookie.Secure = true

	templateManager := templatemanager.NewManager(logger, &isDevelopment)
	cache, err := templateManager.CreateTemplateCache()

	if cache == nil {
		panic("Failed to initialise cache")
	} else if err != nil {
		panic(err)
	}

	templateService := NewTemplateService(templateManager, sessionManager, logger, &isDevelopment)
	formDecoder := form.NewDecoder()

	watcher, err := NewWatcher("./ui/html", templateManager, templateService)
	if err != nil {
		logger.Error("failed to initialise watcher", "error", err)
	}

	return &Services{
		DB:            db,
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
		Validator:     validator.New(db.DB),
	}
}
