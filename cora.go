package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/shaharia-lab/cora/pkg/concatenator"
	"github.com/spf13/cobra"
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

	rootCmd.Flags().StringVarP(&cfg.SourceDirectory, "source", "s", "", "Source directory to Concatenate files from")
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

	debugLog := concatenator.NewDebugLog(cfg.EnableDebugging)
	w := newWalker(cfg.SourceDirectory, cfg.ExcludePatterns, cfg.IncludePatterns, debugLog)
	c := concatenator.NewConcatenation(cfg.OutputFile, cfg.Separator, cfg.PathPrefix, debugLog)

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

func (cfg *config) process(w *walker, c *concatenator.Concatenator) error {
	filePaths, err := w.walk()
	if err != nil {
		return fmt.Errorf("failed to walk directory: %w", err)
	}

	if err := c.Concatenate(filePaths); err != nil {
		return fmt.Errorf("failed to Concatenate files: %w", err)
	}

	return nil
}

type walker struct {
	sourceDirectory string
	excludePatterns []string
	includePatterns []string
	debugLog        *concatenator.DebugLog
}

func newWalker(sourceDirectory string, excludePatterns, includePatterns []string, debugLog *concatenator.DebugLog) *walker {
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
			w.debugLog.Print(fmt.Sprintf("Excluding %s", path))
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
					w.debugLog.Print(fmt.Sprintf("Including %s", path))
				} else {
					w.debugLog.Print(fmt.Sprintf("Skipping %s (not in include patterns)", path))
				}
			} else {
				files = append(files, path)
				w.debugLog.Print(fmt.Sprintf("Including %s", path))
			}
		}

		return nil
	})

	return files, err
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
