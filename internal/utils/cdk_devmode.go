package utils

import (
	"os"
	"strings"
)

func CdkDevMode() bool {
	return strings.ToLower(os.Getenv("CDK_DEV_MODE")) == "true"
}
