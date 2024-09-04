package concatenator

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

const (
	defaultBufferSize = 64 * 1024 // 64KB
)

// Concatenator concatenates files into a single file.
type Concatenator struct {
	outputPath string
	separator  []byte
	debugLog   *DebugLog
	pathPrefix []byte
}

// NewConcatenator creates a new Concatenator.
func NewConcatenator(outputPath, separator, pathPrefix string, debugLog *DebugLog) *Concatenator {
	return &Concatenator{
		outputPath: outputPath,
		separator:  []byte(separator),
		pathPrefix: []byte(pathPrefix),
		debugLog:   debugLog,
	}
}

// Concatenate concatenates the files specified by filePaths into a single file.
func (c *Concatenator) Concatenate(filePaths []string) error {
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

func (c *Concatenator) appendFileContent(writer *bufio.Writer, filePath string) error {
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

func (c *Concatenator) writeFileHeader(writer *bufio.Writer, filePath string) error {
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

type DebugLog struct {
	enabled bool
}

func NewDebugLog(enabled bool) *DebugLog {
	return &DebugLog{
		enabled: enabled,
	}
}

func (d *DebugLog) Print(message string) {
	if !d.enabled {
		return
	}

	log.Println(message)
}
