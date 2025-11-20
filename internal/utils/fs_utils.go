package utils

import (
	"os"
	"path/filepath"
	"runtime"
)

const AppName string = "conduktor"

// GetConfigDir returns the path to the configuration directory for the application based on the OS conventions.
func GetConfigDir() (string, error) {
	var configDir string

	switch runtime.GOOS {
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configDir = filepath.Join(home, "Library", "Application Support", AppName)
	case "linux":
		// Check XDG_CONFIG_HOME first
		if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
			configDir = filepath.Join(xdgConfig, AppName)
		} else {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			configDir = filepath.Join(home, ".config", AppName)
		}
	case "windows":
		if appData := os.Getenv("APPDATA"); appData != "" {
			configDir = filepath.Join(appData, AppName)
		} else {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			configDir = filepath.Join(home, "AppData", "Roaming", AppName)
		}
	default:
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configDir = filepath.Join(home, "."+AppName)
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}

	return configDir, nil
}

// GetDataDir returns the path to the data directory for the application based on the OS conventions.
func GetDataDir() (string, error) {
	var dataDir string

	switch runtime.GOOS {
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dataDir = filepath.Join(home, "Library", "Application Support", AppName)
	case "linux":
		if xdgData := os.Getenv("XDG_DATA_HOME"); xdgData != "" {
			dataDir = filepath.Join(xdgData, AppName)
		} else {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			dataDir = filepath.Join(home, ".local", "share", AppName)
		}
	case "windows":
		if appData := os.Getenv("APPDATA"); appData != "" {
			dataDir = filepath.Join(appData, AppName)
		} else {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			dataDir = filepath.Join(home, "AppData", "Roaming", AppName)
		}
	default:
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dataDir = filepath.Join(home, "."+AppName)
	}

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return "", err
	}

	return dataDir, nil
}
