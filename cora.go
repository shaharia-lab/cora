package main

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

const (
	defaultBufferSize = 64 * 1024 // 64KB
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	cfg := &config{}

	rootCmd := &cobra.Command{
		Use:   "cora",
		Short: "Concatenate files in a directory into a single file.",
		Long:  `Concatenate files in a directory into a single file.`,
		RunE:  cfg.run,
	}

	rootCmd.Flags().StringVarP(&cfg.SourceDirectory, "source", "s", "", "Source directory to concatenate files from")
	rootCmd.Flags().StringVarP(&cfg.OutputFile, "output", "o", "", "Output file to write concatenated files to")
	rootCmd.Flags().StringSliceVarP(&cfg.ExcludePatterns, "exclude", "e", nil, "Glob patterns to exclude")
	rootCmd.Flags().StringSliceVarP(&cfg.IncludePatterns, "include", "i", nil, "Glob patterns to include")
	rootCmd.Flags().BoolVarP(&cfg.EnableDebugging, "debug", "d", false, "Enable debugging mode")
	rootCmd.Flags().StringVarP(&cfg.Separator, "separator", "p", "\n---\n", "Separator to use between concatenated files")
	rootCmd.Flags().StringVarP(&cfg.PathPrefix, "path-prefix", "x", "## ", "Prefix to add before the path of included files")

	return rootCmd
}

type config struct {
	SourceDirectory string
	OutputFile      string
	ExcludePatterns []string
	IncludePatterns []string
	Separator       string
	PathPrefix      string
	EnableDebugging bool
}

func (cfg *config) run(cmd *cobra.Command, args []string) error {
	if err := cfg.validate(); err != nil {
		return err
	}

	debugLog := newDebugLog(cfg.EnableDebugging)
	w := newWalker(cfg.SourceDirectory, cfg.ExcludePatterns, cfg.IncludePatterns, debugLog)
	c := newConcatenator(cfg.OutputFile, cfg.Separator, cfg.PathPrefix, debugLog)

	return cfg.process(w, c)
}

func (cfg *config) validate() error {
	if cfg.SourceDirectory == "" {
		return fmt.Errorf("source directory is required")
	}
	if cfg.OutputFile == "" {
		return fmt.Errorf("output file is required")
	}
	return nil
}

func (cfg *config) process(w *walker, c *concatenator) error {
	filePaths, err := w.walk()
	if err != nil {
		return fmt.Errorf("failed to walk directory: %w", err)
	}

	if err := c.concatenate(filePaths); err != nil {
		return fmt.Errorf("failed to concatenate files: %w", err)
	}

	return nil
}

type walker struct {
	sourceDirectory string
	excludePatterns []string
	includePatterns []string
	debugLog        *debugLog
}

func newWalker(sourceDirectory string, excludePatterns, includePatterns []string, debugLog *debugLog) *walker {
	return &walker{
		sourceDirectory: sourceDirectory,
		excludePatterns: excludePatterns,
		includePatterns: includePatterns,
		debugLog:        debugLog,
	}
}

func (w *walker) walk() ([]string, error) {
	var files []string

	err := filepath.WalkDir(w.sourceDirectory, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		excluded, err := matchesGlob(w.sourceDirectory, path, w.excludePatterns)
		if err != nil {
			return err
		}

		if excluded {
			w.debugLog.print(fmt.Sprintf("Excluding %s", path))
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if !d.IsDir() {
			if len(w.includePatterns) > 0 {
				included, err := matchesGlob(w.sourceDirectory, path, w.includePatterns)
				if err != nil {
					return err
				}
				if included {
					files = append(files, path)
					w.debugLog.print(fmt.Sprintf("Including %s", path))
				} else {
					w.debugLog.print(fmt.Sprintf("Skipping %s (not in include patterns)", path))
				}
			} else {
				files = append(files, path)
				w.debugLog.print(fmt.Sprintf("Including %s", path))
			}
		}

		return nil
	})

	return files, err
}

type concatenator struct {
	outputPath string
	separator  []byte
	debugLog   *debugLog
	pathPrefix []byte
}

func newConcatenator(outputPath, separator, pathPrefix string, debugLog *debugLog) *concatenator {
	return &concatenator{
		outputPath: outputPath,
		separator:  []byte(separator),
		pathPrefix: []byte(pathPrefix),
		debugLog:   debugLog,
	}
}

func (c *concatenator) concatenate(filePaths []string) error {
	if err := os.MkdirAll(filepath.Dir(c.outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	outFile, err := os.Create(c.outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	writer := bufio.NewWriterSize(outFile, defaultBufferSize)
	defer writer.Flush()

	for i, filePath := range filePaths {
		if i > 0 {
			if _, err := writer.Write(c.separator); err != nil {
				return fmt.Errorf("failed to write separator: %w", err)
			}
		}

		if err := c.writeFileHeader(writer, filePath); err != nil {
			return err
		}

		if err := c.appendFileContent(writer, filePath); err != nil {
			return err
		}

		if _, err := writer.Write([]byte{'\n'}); err != nil {
			return fmt.Errorf("failed to write newline after file content: %w", err)
		}
	}

	return nil
}

func (c *concatenator) writeFileHeader(writer *bufio.Writer, filePath string) error {
	if _, err := writer.Write(c.pathPrefix); err != nil {
		return fmt.Errorf("failed to write path prefix: %w", err)
	}

	if _, err := writer.WriteString(filePath); err != nil {
		return fmt.Errorf("failed to write file path: %w", err)
	}

	if _, err := writer.Write([]byte{'\n'}); err != nil {
		return fmt.Errorf("failed to write newline after file path: %w", err)
	}

	return nil
}

func (c *concatenator) appendFileContent(writer *bufio.Writer, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	_, err = io.Copy(writer, file)
	if err != nil {
		return fmt.Errorf("failed to copy content from %s: %w", filePath, err)
	}

	return nil
}

func matchesGlob(rootPath, filePath string, patterns []string) (bool, error) {
	relPath, err := filepath.Rel(rootPath, filePath)
	if err != nil {
		return false, err
	}

	relPath = filepath.ToSlash(relPath)

	for _, pattern := range patterns {
		pattern = filepath.ToSlash(pattern)

		if !strings.Contains(pattern, "/") {
			matched, err := filepath.Match(pattern, filepath.Base(relPath))
			if err != nil {
				return false, err
			}
			if matched {
				return true, nil
			}
		} else {
			matched, err := filepath.Match(pattern, relPath)
			if err != nil {
				return false, err
			}
			if matched {
				return true, nil
			}
		}
	}

	return false, nil
}

type debugLog struct {
	enabled bool
}

func newDebugLog(enabled bool) *debugLog {
	return &debugLog{
		enabled: enabled,
	}
}

func (d *debugLog) print(message string) {
	if !d.enabled {
		return
	}

	log.Println(message)
}
