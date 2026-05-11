package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
)

func main() {
	dirs := []string{"internal/service", "internal/handler"}

	// Regex for date parsing
	reDate := regexp.MustCompile(`time\.Parse\(\s*"2006-01-02"\s*,\s*([^)]+)\)`)
	// Regex for time parsing
	reTime := regexp.MustCompile(`time\.Parse\(\s*"15:04"\s*,\s*([^)]+)\)`)
	// Regex for layout parsing
	reLayout := regexp.MustCompile(`time\.Parse\(\s*(layout)\s*,\s*([^)]+)\)`)

	// Regex for parsing exact times
	reFullTime := regexp.MustCompile(`time\.Parse\(\s*"2006-01-02 15:04:05"\s*,\s*([^)]+)\)`)
	reFullTime2 := regexp.MustCompile(`time\.Parse\(\s*"2006-01"\s*,\s*([^)]+)\)`)

	for _, dir := range dirs {
		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && filepath.Ext(path) == ".go" {
				content, err := ioutil.ReadFile(path)
				if err != nil {
					return err
				}

				origContent := content

				// Replace
				content = reDate.ReplaceAll(content, []byte(`utils.ParseDateWIB($1)`))
				content = reTime.ReplaceAll(content, []byte(`utils.ParseTimeWIB("15:04", $1)`))
				content = reLayout.ReplaceAll(content, []byte(`utils.ParseTimeWIB($1, $2)`))
				content = reFullTime.ReplaceAll(content, []byte(`utils.ParseTimeWIB("2006-01-02 15:04:05", $1)`))
				content = reFullTime2.ReplaceAll(content, []byte(`utils.ParseTimeWIB("2006-01", $1)`))

				if !bytes.Equal(origContent, content) {
					// Ensure utils import is added. A simple hack is to just add it; goimports will clean it up.
					// Or just let goimports add it. Let's see if goimports is available. We'll run it in shell.
					err = ioutil.WriteFile(path, content, 0644)
					if err != nil {
						fmt.Printf("Error writing %s: %v\n", path, err)
					} else {
						fmt.Printf("Updated %s\n", path)
					}
				}
			}
			return nil
		})
	}
}
