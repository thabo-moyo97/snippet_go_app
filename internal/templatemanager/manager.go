package templatemanager

import "strings"

type Manager struct {
	extensions []string
}

func NewManager() *Manager {
	return &Manager{
		extensions: []string{".tmpl", ".html", ".phtml"},
	}
}

func (m *Manager) SetExtensions(exts []string) {
	m.extensions = exts
}

func (m *Manager) NormaliseTemplateName(name string) string {
	name = strings.TrimPrefix(name, "ui/html/pages/")
	name = strings.TrimPrefix(name, "html/pages/")
	name = strings.TrimPrefix(name, "html/partials/")
	name = strings.TrimPrefix(name, "html/")

	name = strings.ReplaceAll(name, "/", ".")

	for _, ext := range m.extensions {
		name = strings.TrimSuffix(name, ext)
	}

	return name
}

func (m *Manager) GetExtensions() []string {
	return m.extensions
}
