package utils

func TextOf[K comparable](m map[K]string, k K) string {
	if v, ok := m[k]; ok {
		return v
	}
	return "未知"
}
