// internal/utils/preload.go
package utils

import "gorm.io/gorm"

func ApplyPreloads(query *gorm.DB, includes []string, preloadMap map[string]string) *gorm.DB {
	for _, include := range includes {
		if rel, ok := preloadMap[include]; ok {
			query = query.Preload(rel)
		}
	}
	return query
}
