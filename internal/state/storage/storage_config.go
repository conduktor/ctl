package storage

type StorageConfig struct {
	StateFilePath *string
}

func NewStorageConfig(stateFilePath *string) StorageConfig {
	return StorageConfig{
		StateFilePath: stateFilePath,
	}
}
