package utils

import "strings"

func DesensitizeMobile(mobile string) string {
	if len(mobile) != 11 {
		return mobile
	}
	return mobile[:3] + "****" + mobile[7:]
}

func DisplayName(nickname string) string {
	return strings.TrimSpace(nickname)
}
