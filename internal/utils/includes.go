package utils

import "strings"

func ParseIncludes(input string) []string {
	var result []string

	if input == "" {
		return result
	}

	parts := strings.Split(input, ",")

	for _, part := range parts {
		val := strings.TrimSpace(part)
		if val != "" {
			result = append(result, val)
		}
	}

	return result
}
