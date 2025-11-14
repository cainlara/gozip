package util

import (
	"archive/zip"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cainlara/gozip/core"
)

func GetFileToExtract() (string, []core.ZippedFile, error) {
	execFolder, err := getExecutionFolder()
	if err != nil {
		return "", nil, err
	}

	fileName, err := getFileArgumentValue()
	if err != nil {
		return "", nil, err
	}

	filePath := filepath.Join(execFolder, fileName)

	content, err := openZipFile(filePath)
	if err != nil {
		return "", nil, err
	}

	return fileName, content, nil
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
