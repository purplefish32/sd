package util

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

func CreateLockFile(filename string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	sdDir := filepath.Join(home, ".streamdeck")
	if err := os.MkdirAll(sdDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	lockFile := filepath.Join(sdDir, filename)
	file, err := os.OpenFile(lockFile, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("failed to open lock file: %w", err)
	}

	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		file.Close()
		return fmt.Errorf("another instance is already running")
	}

	return nil
}
