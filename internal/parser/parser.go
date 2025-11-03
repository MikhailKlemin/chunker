package parser

import (
	"os"
	"path/filepath"
	"strings"
)

// FindCppFiles recursively finds all C++ source and header files
func FindCppFiles(rootPath string) ([]string, error) {
	var files []string

	// Directories to skip
	skipDirs := map[string]bool{
		"build":        true,
		".git":         true,
		"node_modules": true,
		"CMakeFiles":   true,
		".cache":       true,
		"vendor":       true,
	}

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Log but continue walking
			return nil
		}

		// Skip excluded directories
		if info.IsDir() {
			if skipDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		// Check for C++ file extensions
		if isCppFile(path) {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

// isCppFile checks if a file has a C++ extension
func isCppFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))

	cppExtensions := []string{
		".cpp", ".cc", ".cxx", ".c++", // C++ source files
		".hpp", ".h", ".hxx", ".h++", // C++ header files
	}

	for _, validExt := range cppExtensions {
		if ext == validExt {
			return true
		}
	}

	return false
}

// GetFileStats returns statistics about discovered files
func GetFileStats(files []string) map[string]int {
	stats := make(map[string]int)

	for _, file := range files {
		ext := strings.ToLower(filepath.Ext(file))
		stats[ext]++
	}

	return stats
}
