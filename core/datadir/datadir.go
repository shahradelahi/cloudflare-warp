package datadir

import (
	"os"
	"path"
)

var dataDir string

// SetDataDir sets the data directory path.
func SetDataDir(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0660); err != nil {
			panic(err)
		}
	}

	dataDir = dir
}

// GetDataDir returns the data directory path.
func GetDataDir() string {
	return dataDir
}

// GetDataDirOrPath determines the data directory path.
// It checks the provided dataDir, then the HOME environment variable,
// and finally defaults to ".cloudflare-warp".
func GetDataDirOrPath(dir string) string {
	switch {
	case dir != "":
		return dir
	case os.Getenv("HOME") != "":
		return path.Join(os.Getenv("HOME"), ".cloudflare-warp")
	default:
		return ".cloudflare-warp"
	}
}
