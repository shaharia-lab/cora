package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/shaharia-lab/cora/pkg/concatenator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConcatenator(t *testing.T) {
	tempDir := t.TempDir()
	inputFiles := []string{
		filepath.Join(tempDir, "file1.txt"),
		filepath.Join(tempDir, "file2.txt"),
		filepath.Join(tempDir, "file3.txt"),
	}
	outputFile := filepath.Join(tempDir, "output.txt")

	for i, file := range inputFiles {
		err := os.WriteFile(file, []byte(fmt.Sprintf("Content of file %d", i+1)), 0644)
		require.NoError(t, err)
	}

	debugLog := concatenator.NewDebugLog(false)
	conc := concatenator.NewConcatenation(outputFile, "\n---\n", "File: ", debugLog)
	err := conc.Concatenate(inputFiles)
	require.NoError(t, err)

	content, err := os.ReadFile(outputFile)
	require.NoError(t, err)

	var expectedContent bytes.Buffer
	for i, file := range inputFiles {
		if i > 0 {
			expectedContent.WriteString("\n---\n")
		}
		expectedContent.WriteString(fmt.Sprintf("File: %s\n", file))
		expectedContent.WriteString(fmt.Sprintf("Content of file %d\n", i+1))
	}

	assert.Equal(t, expectedContent.String(), string(content))
}

func TestConcatenatorLargeFiles(t *testing.T) {
	tempDir := t.TempDir()
	inputFiles := make([]string, 100)
	for i := 0; i < 100; i++ {
		inputFiles[i] = filepath.Join(tempDir, fmt.Sprintf("file%d.txt", i))
	}
	outputFile := filepath.Join(tempDir, "output.txt")

	largeContent := bytes.Repeat([]byte("a"), 1024*1024)
	for _, file := range inputFiles {
		err := os.WriteFile(file, largeContent, 0644)
		require.NoError(t, err)
	}

	debugLog := concatenator.NewDebugLog(false)
	conc := concatenator.NewConcatenation(outputFile, "\n", "File: ", debugLog)
	err := conc.Concatenate(inputFiles)
	require.NoError(t, err)

	stat, err := os.Stat(outputFile)
	require.NoError(t, err)

	filePathLen := len(inputFiles[0])
	prefixLen := len("File: ")
	separatorLen := len("\n")
	singleFileSize := int64(1024*1024 + filePathLen + prefixLen + 2)
	expectedSize := 100*singleFileSize + 99*int64(separatorLen) + 90

	content, err := os.ReadFile(outputFile)
	require.NoError(t, err)

	assert.Equal(t, expectedSize, stat.Size())
	assert.Equal(t, int(expectedSize), len(content))
}

func TestWalker(t *testing.T) {
	root := t.TempDir()
	createTestFiles(t, root)

	debugLog := concatenator.NewDebugLog(false)

	tests := []struct {
		name            string
		excludePatterns []string
		includePatterns []string
		expectedFiles   []string
	}{
		{"No patterns", []string{}, []string{}, []string{"file1.txt", "file2.txt", "file3.txt", "ignoreme.txt"}},
		{"Exclude one dir", []string{"ignoreme"}, []string{}, []string{"file1.txt", "file2.txt", "file3.txt"}},
		{"Include specific files", []string{}, []string{"file1.txt", "file3.txt"}, []string{"file1.txt", "file3.txt"}},
		{"Include and exclude", []string{"ignoreme"}, []string{"file*.txt"}, []string{"file1.txt", "file2.txt", "file3.txt"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			walker := newWalker(root, tt.excludePatterns, tt.includePatterns, debugLog)
			files, err := walker.walk()

			assert.NoError(t, err)

			var baseNames []string
			for _, file := range files {
				baseNames = append(baseNames, filepath.Base(file))
			}

			assert.ElementsMatch(t, tt.expectedFiles, baseNames)
		})
	}
}

func createTestFiles(t *testing.T, root string) {
	files := []string{
		"file1.txt",
		"file2.txt",
		"file3.txt",
		filepath.Join("ignoreme", "ignoreme.txt"),
	}

	for _, file := range files {
		path := filepath.Join(root, file)
		err := os.MkdirAll(filepath.Dir(path), 0755)
		assert.NoError(t, err)

		err = os.WriteFile(path, []byte("test"), 0644)
		assert.NoError(t, err)
	}
}

func TestMatchesGlob(t *testing.T) {
	tests := []struct {
		name     string
		rootPath string
		filePath string
		patterns []string
		expected bool
	}{
		{
			name:     "Match single file",
			rootPath: "/root",
			filePath: "/root/file.txt",
			patterns: []string{"*.txt"},
			expected: true,
		},
		{
			name:     "Match file in subdirectory",
			rootPath: "/root",
			filePath: "/root/subdir/file.go",
			patterns: []string{"**/*.go"},
			expected: true,
		},
		{
			name:     "No match",
			rootPath: "/root",
			filePath: "/root/file.txt",
			patterns: []string{"*.go"},
			expected: false,
		},
		{
			name:     "Match with multiple patterns",
			rootPath: "/root",
			filePath: "/root/subdir/file.js",
			patterns: []string{"*.go", "**/*.js"},
			expected: true,
		},
		{
			name:     "Match file name only",
			rootPath: "/root",
			filePath: "/root/subdir/config.json",
			patterns: []string{"config.json"},
			expected: true,
		},
		{
			name:     "Match directory name",
			rootPath: "/root",
			filePath: "/root/ignoreme",
			patterns: []string{"ignoreme"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := matchesGlob(tt.rootPath, tt.filePath, tt.patterns)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDebugLog(t *testing.T) {
	tests := []struct {
		name    string
		enabled bool
		message string
	}{
		{"Enabled", true, "Test message"},
		{"Disabled", false, "Test message"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			debugLog := concatenator.NewDebugLog(tt.enabled)

			// Capture log output
			var buf bytes.Buffer
			log.SetOutput(&buf)
			defer log.SetOutput(os.Stderr)

			debugLog.Print(tt.message)

			if tt.enabled {
				assert.Contains(t, buf.String(), tt.message)
			} else {
				assert.Empty(t, buf.String())
			}
		})
	}
}
