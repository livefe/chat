package utils

import (
	"os"
	"strings"

	"github.com/google/uuid"
)

func InstanceID() string {
	host, err := os.Hostname()
	if err == nil {
		if host = strings.TrimSpace(host); host != "" {
			return host
		}
	}
	return uuid.NewString()
}
