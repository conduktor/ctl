package utils

import (
	"os"
	"strings"
)

func CdkDebug() bool {
	return strings.ToLower(os.Getenv("CDK_DEBUG")) == "true"
}
