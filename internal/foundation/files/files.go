package files

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ReadSQLFiles reads all .sql files from a directory, ordered by name,
// and returns a slice with the content of each file.
func ReadSQLFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	var contents []string
	for _, name := range names {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", name, err)
		}
		contents = append(contents, string(data))
	}

	return contents, nil
}
