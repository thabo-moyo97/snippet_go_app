package ui

import (
	"io/fs"
	"os"
)

// Files provides direct filesystem access during local development
var ViewFiles fs.FS = os.DirFS("ui")
