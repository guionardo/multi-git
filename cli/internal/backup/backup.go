package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Backup copies src to src.bak.<timestamp> if it exists.
// Returns backup path (empty if no source).
func Backup(src string) (string, error) {
	if _, err := os.Stat(src); err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	ts := time.Now().Format("20060102-150405")
	dst := src + ".bak." + ts
	in, err := os.ReadFile(src)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return "", err
	}
	if err := os.WriteFile(dst, in, 0644); err != nil {
		return "", err
	}
	return dst, nil
}

func Restore(src, backup string) error {
	b, err := os.ReadFile(backup)
	if err != nil {
		return err
	}
	return os.WriteFile(src, b, 0644)
}

func Exists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

var _ = fmt.Sprintf
