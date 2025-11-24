package storage

import (
	"os"
)

type StorageConfig struct {
	Enabled  bool
	FilePath *string
}

// NewStorageConfig creates a StorageConfig based on the provided pointers.
// If the pointers are nil, it falls back to environment variables:
// - CDK_STATE_ENABLED for Enabled (expects "true" or "false")
// - CDK_STATE_FILE for FilePath.
func NewStorageConfig(stateEnabled *bool, stateFilePath *string) StorageConfig {
	enable := false
	if stateEnabled == nil || *stateEnabled == false {
		enabledEnv := os.Getenv("CDK_STATE_ENABLED")
		if enabledEnv == "true" || enabledEnv == "1" || enabledEnv == "yes" {
			enable = true
		}
	} else {
		enable = *stateEnabled
	}

	var filePath *string
	if stateFilePath == nil || *stateFilePath == "" {
		filePathEnv := os.Getenv("CDK_STATE_FILE")
		if filePathEnv != "" {
			filePath = &filePathEnv
		}
	} else {
		filePath = stateFilePath
	}

	return StorageConfig{
		Enabled:  enable,
		FilePath: filePath,
	}
}
