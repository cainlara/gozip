// Package util provides utility functions for working with ZIP files.
// It includes functionality for parsing command-line arguments,
// opening ZIP files, and extracting their contents.
package util

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cainlara/gozip/core"
)

// GetFileToExtract retrieves the ZIP file specified in command-line arguments
// and extracts its contents.
//
// This function performs the following operations:
//  1. Obtains the current execution directory
//  2. Parses command-line arguments to get the ZIP file name
//  3. Constructs the full path to the ZIP file
//  4. Opens and reads the contents of the ZIP file
//
// Returns:
//   - string: name of the ZIP file
//   - string: full path to the ZIP file
//   - []core.ZippedFile: slice containing all files within the ZIP
//   - error: any error encountered during the process
//
// Possible errors:
//   - Error obtaining the execution directory
//   - Error parsing arguments (no arguments, too many arguments, invalid extension)
//   - Error opening the ZIP file (file doesn't exist, not a valid ZIP)
func GetFileToExtract() (string, string, []core.ZippedFile, error) {
	execFolder, err := getExecutionFolder()
	if err != nil {
		return "", "", nil, err
	}

	fileName, err := getFileArgumentValue()
	if err != nil {
		return "", "", nil, err
	}

	filePath := filepath.Join(execFolder, fileName)

	content, err := openZipFile(filePath)
	if err != nil {
		return "", "", nil, err
	}

	return fileName, filePath, content, nil
}

func getExecutionFolder() (string, error) {
	ex, err := os.Getwd()

	if err != nil {
		return "", err
	}

	return ex, nil
}

func getFileArgumentValue() (string, error) {
	args := os.Args
	if len(args) > 2 {
		return "", errors.New("i don't know what to do with so many arguments")
	}

	argsWithoutProg := args[1:]
	if len(argsWithoutProg) == 0 {
		return "", errors.New("no zip file provided")
	}

	fileName := argsWithoutProg[0]

	if len(fileName) == 0 || !strings.HasSuffix(fileName, ".zip") {
		return "", errors.New("invalid zip file name")
	}

	return fileName, nil
}

func openZipFile(filePath string) ([]core.ZippedFile, error) {
	reader, err := zip.OpenReader(filePath)
	if err != nil {
		return nil, err
	}

	defer reader.Close()

	content := make([]core.ZippedFile, 0, len(reader.File))

	for _, f := range reader.File {
		fi := f.FileInfo()
		name := f.Name
		isDir := fi.IsDir()
		uncompressed := f.UncompressedSize64
		compressed := f.CompressedSize64
		method := methodToString(f.Method)

		var modStr string
		if !f.Modified.IsZero() {
			modStr = f.Modified.UTC().Format(time.RFC3339)
		} else {
			modStr = "-"
		}

		crc := f.CRC32

		zf := core.NewZippedFile(name, isDir, uncompressed, compressed, method, modStr, crc)
		content = append(content, zf)
	}

	return content, nil
}

func methodToString(m uint16) string {
	switch m {
	case 0:
		return "STORE"
	case 8:
		return "DEFLATE"
	default:
		return fmt.Sprintf("0x%X", m)
	}
}

// ExtractFile extracts a file or folder from a ZIP archive to the destination directory.
//
// If the target is a file, only that file is extracted.
// If the target is a folder, all contents within that folder are extracted recursively.
//
// Parameters:
//   - zipPath: full path to the ZIP file
//   - targetName: name of the file or folder to extract (as it appears in the ZIP)
//   - destDir: destination directory where files will be extracted
//
// Returns:
//   - int: number of files extracted
//   - error: any error encountered during extraction
func ExtractFile(zipPath, targetName, destDir string) (int, error) {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return 0, fmt.Errorf("failed to open ZIP file: %w", err)
	}
	defer reader.Close()

	// Normalize target name to handle both files and folders
	targetPrefix := targetName
	if !strings.HasSuffix(targetPrefix, "/") {
		targetPrefix = targetName + "/"
	}

	var extractedCount int
	var found bool

	for _, f := range reader.File {
		// Check if this file matches (exact match) or is within the target folder
		if f.Name == targetName || strings.HasPrefix(f.Name, targetPrefix) {
			found = true

			// Skip if it's a directory entry
			if f.FileInfo().IsDir() {
				continue
			}

			// Construct destination path
			destPath := filepath.Join(destDir, f.Name)

			// Create parent directories
			if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
				return extractedCount, fmt.Errorf("failed to create directory: %w", err)
			}

			// Extract the file
			if err := extractSingleFile(f, destPath); err != nil {
				return extractedCount, fmt.Errorf("failed to extract %s: %w", f.Name, err)
			}

			extractedCount++
		}
	}

	if !found {
		return 0, fmt.Errorf("file or folder '%s' not found in ZIP archive", targetName)
	}

	return extractedCount, nil
}

// extractSingleFile extracts a single file from the ZIP archive to the destination path.
func extractSingleFile(f *zip.File, destPath string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	outFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, rc)
	return err
}
