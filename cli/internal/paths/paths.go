package paths

import (
	"os"
	"path/filepath"
	"strings"
)

func Home() string {
	h, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return h
}

func Join(elem ...string) string {
	return filepath.Join(elem...)
}

// GitPath returns a path using forward slashes (preferred by Git on all OSes).
func GitPath(p string) string {
	return filepath.ToSlash(p)
}

func ExpandHome(p string) string {
	if strings.HasPrefix(p, "~/") {
		return filepath.Join(Home(), p[2:])
	}
	return p
}