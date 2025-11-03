package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindCppFiles(t *testing.T) {
	// Create temporary test directory
	tmpDir, err := os.MkdirTemp("", "clangd-parser-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test file structure
	testFiles := map[string]bool{
		"main.cpp":            true,  // Should find
		"helper.cc":           true,  // Should find
		"utils.cxx":           true,  // Should find
		"header.hpp":          true,  // Should find
		"header.h":            true,  // Should find
		"readme.txt":          false, // Should skip
		"config.json":         false, // Should skip
		"src/module.cpp":      true,  // Should find (nested)
		"src/module.h":        true,  // Should find (nested)
		"build/temp.cpp":      false, // Should skip (build dir)
		".git/config":         false, // Should skip (.git dir)
		"CMakeFiles/test.cpp": false, // Should skip (CMakeFiles dir)
	}

	// Create the test files
	for filePath := range testFiles {
		fullPath := filepath.Join(tmpDir, filePath)

		// Create parent directory if needed
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create dir %s: %v", dir, err)
		}

		// Create the file
		if err := os.WriteFile(fullPath, []byte("// test"), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", fullPath, err)
		}
	}

	// Run FindCppFiles
	files, err := FindCppFiles(tmpDir)
	if err != nil {
		t.Fatalf("FindCppFiles failed: %v", err)
	}

	// Create a map of found files for easy lookup
	foundFiles := make(map[string]bool)
	for _, file := range files {
		relPath, _ := filepath.Rel(tmpDir, file)
		foundFiles[relPath] = true
	}

	// Verify results
	for filePath, shouldFind := range testFiles {
		if shouldFind {
			if !foundFiles[filePath] {
				t.Errorf("Expected to find %s but didn't", filePath)
			}
		} else {
			if foundFiles[filePath] {
				t.Errorf("Found %s but shouldn't have", filePath)
			}
		}
	}

	t.Logf("✓ Found %d C++ files (expected 7)", len(files))
}

func TestIsCppFile(t *testing.T) {
	tests := []struct {
		filename string
		expected bool
	}{
		{"main.cpp", true},
		{"helper.cc", true},
		{"utils.cxx", true},
		{"advanced.c++", true},
		{"header.hpp", true},
		{"header.h", true},
		{"header.hxx", true},
		{"header.h++", true},
		{"Main.CPP", true},
		{"Header.HPP", true},
		{"readme.txt", false},
		{"config.json", false},
		{"script.py", false},
		{"data.csv", false},
	}

	for _, tt := range tests {
		result := isCppFile(tt.filename)
		if result != tt.expected {
			t.Errorf("isCppFile(%q) = %v, expected %v", tt.filename, result, tt.expected)
		}
	}

	t.Log("✓ isCppFile tests passed")
}

func TestGetFileStats(t *testing.T) {
	files := []string{
		"main.cpp",
		"helper.cpp",
		"utils.cc",
		"header.hpp",
		"header.h",
		"another.h",
	}

	stats := GetFileStats(files)

	expected := map[string]int{
		".cpp": 2,
		".cc":  1,
		".hpp": 1,
		".h":   2,
	}

	for ext, count := range expected {
		if stats[ext] != count {
			t.Errorf("Expected %d files with extension %s, got %d", count, ext, stats[ext])
		}
	}

	t.Logf("✓ File stats: %v", stats)
}

func TestFindCppFilesEmptyDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "clangd-parser-empty-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	files, err := FindCppFiles(tmpDir)
	if err != nil {
		t.Fatalf("FindCppFiles failed: %v", err)
	}

	if len(files) != 0 {
		t.Errorf("Expected 0 files in empty directory, got %d", len(files))
	}

	t.Log("✓ Empty directory test passed")
}

func TestFindCppFilesNonExistentDirectory(t *testing.T) {
	files, err := FindCppFiles("/nonexistent/directory/that/does/not/exist")

	// Should handle gracefully
	if err != nil {
		t.Logf("✓ Correctly handled non-existent directory: %v", err)
	} else if len(files) == 0 {
		t.Log("✓ Returned empty list for non-existent directory")
	}
}
