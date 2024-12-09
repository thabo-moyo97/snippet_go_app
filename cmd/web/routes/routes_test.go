package routes

import (
	"testing"

	"thabomoyo.co.uk/ui"
)

func TestEmbeddedFiles(t *testing.T) {
	// Try to read the main.css file
	data, err := ui.Files.ReadFile("static/css/main.css")
	if err != nil {
		t.Errorf("Failed to read main.css: %v", err)
	}
	if len(data) == 0 {
		t.Error("main.css is empty")
	}
}

func TestStaticFiles(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
	}{
		{
			name:     "CSS file",
			filePath: "static/css/main.css",
		},
		{
			name:     "Favicon",
			filePath: "static/img/favicon.ico",
		},
		{
			name:     "Logo",
			filePath: "static/img/logo.png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := ui.Files.ReadFile(tt.filePath)
			if err != nil {
				t.Errorf("Failed to read %s: %v", tt.filePath, err)
			}
			if len(data) == 0 {
				t.Errorf("%s is empty", tt.filePath)
			}
		})
	}
}
