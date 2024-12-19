package services

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
	"thabomoyo.co.uk/internal/templatemanager"
)

type Watcher struct {
	watcher   *fsnotify.Watcher
	clients   []*websocket.Conn
	templates map[string]string // Maps file paths to template names
	tmplSvc   *TemplateService
}

type ReloadMessage struct {
	Type     string `json:"type"`
	Partial  bool   `json:"partial"`
	Template string `json:"template,omitempty"`
}

func NewWatcher(templateDir string, templateManager *templatemanager.Manager, templateService *TemplateService) (*Watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	w := &Watcher{
		watcher:   fsWatcher,
		clients:   make([]*websocket.Conn, 0),
		templates: make(map[string]string),
		tmplSvc:   templateService,
	}

	err = filepath.Walk(templateDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".html" {
			// Map file paths to template names
			templateName := templateManager.NormaliseTemplateName(path)
			w.templates[path] = templateName
			return w.watcher.Add(path)
		}
		return nil
	})

	go w.watch(templateManager)
	return w, err
}

func (w *Watcher) watch(templateManager *templatemanager.Manager) {
	for {
		select {
		case event := <-w.watcher.Events:
			if event.Op&fsnotify.Write == fsnotify.Write {
				templateName := w.templates[event.Name]
				templateName = templateManager.NormaliseTemplateName(templateName)

				msg := ReloadMessage{
					Type:     "reload",
					Partial:  templateName != "base",
					Template: templateName,
				}

				_, err := json.Marshal(msg)
				if err != nil {
					w.tmplSvc.logger.Error("error marshaling reload message", "error", err)
					continue
				}
				if len(w.clients) == 0 {
					w.tmplSvc.logger.Info("no clients connected")
				}
				w.SendMessage(msg)
			}
		case err := <-w.watcher.Errors:
			w.tmplSvc.logger.Error("watcher error", "error", err)
		}
	}
}

func (w *Watcher) AddClient(conn *websocket.Conn) {
	w.clients = append(w.clients, conn)
	w.tmplSvc.logger.Info("client added", "addr", conn.RemoteAddr())
}

func (w *Watcher) RemoveClient(conn *websocket.Conn) {
	for i, client := range w.clients {
		if client == conn {
			w.clients = append(w.clients[:i], w.clients[i+1:]...)
			w.tmplSvc.logger.Info("client removed", "addr", conn.RemoteAddr())
			break
		}
	}
}

func (w *Watcher) SendMessage(msg ReloadMessage) {
	for _, client := range w.clients {
		client.WriteJSON(msg)
	}
}
