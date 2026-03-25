package main

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNoDirectDatabaseImports ensures that only main.go imports the database
// package. All other packages must receive their *gorm.DB via dependency
// injection (constructor parameters). This prevents regression to the old
// pattern of scattering database.DB calls throughout handlers and services.
func TestNoDirectDatabaseImports(t *testing.T) {
	forbiddenImport := "github.com/briandenicola/ancient-coins-api/database"

	// Directories that must NOT import the database package directly
	restricted := []string{
		"handlers",
		"services",
		"middleware",
		"repository",
	}

	for _, dir := range restricted {
		dirPath := filepath.Join(".", dir)
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			continue
		}

		entries, err := os.ReadDir(dirPath)
		if err != nil {
			t.Fatalf("Failed to read directory %s: %v", dir, err)
		}

		fset := token.NewFileSet()
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
				continue
			}

			filePath := filepath.Join(dirPath, entry.Name())
			f, err := parser.ParseFile(fset, filePath, nil, parser.ImportsOnly)
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", filePath, err)
			}

			for _, imp := range f.Imports {
				importPath := strings.Trim(imp.Path.Value, `"`)
				if importPath == forbiddenImport {
					t.Errorf(
						"%s imports %q directly. Use dependency injection instead — "+
							"accept *gorm.DB or a repository as a constructor parameter. "+
							"Only main.go should reference the database package.",
						filePath, forbiddenImport,
					)
				}
			}
		}
	}
}

// TestHandlersDoNotUseRawSQL checks that handler files do not contain raw SQL
// query strings, which should live in the repository layer.
func TestHandlersDoNotUseRawSQL(t *testing.T) {
	handlersDir := filepath.Join(".", "handlers")
	if _, err := os.Stat(handlersDir); os.IsNotExist(err) {
		t.Skip("handlers directory not found")
	}

	sqlPatterns := []string{
		"SELECT ",
		"INSERT INTO",
		"UPDATE ",
		"DELETE FROM",
		".Raw(",
		".Exec(",
	}

	// These are false-positive patterns that appear in non-SQL contexts
	allowList := []string{
		"swagger_types.go",
	}

	entries, err := os.ReadDir(handlersDir)
	if err != nil {
		t.Fatalf("Failed to read handlers directory: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}

		allowed := false
		for _, a := range allowList {
			if entry.Name() == a {
				allowed = true
				break
			}
		}
		if allowed {
			continue
		}

		filePath := filepath.Join(handlersDir, entry.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("Failed to read %s: %v", filePath, err)
		}

		lines := strings.Split(string(content), "\n")
		for lineNum, line := range lines {
			trimmed := strings.TrimSpace(line)
			// Skip comments and string constants (prompts, etc.)
			if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") {
				continue
			}

			for _, pattern := range sqlPatterns {
				if strings.Contains(line, pattern) {
					// Ignore if inside a backtick or quote string constant (prompts contain SQL-like words)
					if isInsideStringConstant(lines, lineNum) {
						continue
					}
					t.Errorf(
						"%s:%d contains raw SQL pattern %q. "+
							"SQL queries belong in the repository layer, not handlers.",
						filePath, lineNum+1, pattern,
					)
				}
			}
		}
	}
}

// isInsideStringConstant is a heuristic to check if a line is inside a
// multi-line backtick string (used for prompts). It counts backticks
// before the line — an odd count means we're inside a raw string.
func isInsideStringConstant(lines []string, targetLine int) bool {
	backtickCount := 0
	for i := 0; i < targetLine; i++ {
		backtickCount += strings.Count(lines[i], "`")
	}
	return backtickCount%2 == 1
}
