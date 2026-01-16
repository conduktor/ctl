package storage

import (
	"os"
)

type StorageConfig struct {
	Enabled   bool
	FilePath  *string
	RemoteURI *string
}

// NewStorageConfig creates a StorageConfig based on the provided pointers.
// If the pointers are nil, it falls back to environment variables:
// - CDK_STATE_ENABLED for Enabled (expects "true" or "false")
// - CDK_STATE_FILE for FilePath (local backend)
// - CDK_STATE_REMOTE_URI for RemoteURI (remote backend)
//
// If RemoteURI is provided, the remote backend will be used.
// Otherwise, the local file backend will be used.
func NewStorageConfig(stateEnabled *bool, stateFilePath *string, stateRemoteURI *string) StorageConfig {
	enable := false
	if stateEnabled == nil || !*stateEnabled {
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

	var remoteURI *string
	if stateRemoteURI == nil || *stateRemoteURI == "" {
		remoteURIEnv := os.Getenv("CDK_STATE_REMOTE_URI")
		if remoteURIEnv != "" {
			remoteURI = &remoteURIEnv
		}
	} else {
		remoteURI = stateRemoteURI
	}

	return StorageConfig{
		Enabled:   enable,
		FilePath:  filePath,
		RemoteURI: remoteURI,
	}
}
