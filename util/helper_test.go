package util

import (
	"os"
	"testing"
)

// TestMethodToString verifies that the conversion of method codes
// to strings works correctly for standard compression methods
func TestMethodToString(t *testing.T) {
	tests := []struct {
		name     string
		method   uint16
		expected string
	}{
		{
			name:     "STORE method (no compression)",
			method:   0,
			expected: "STORE",
		},
		{
			name:     "DEFLATE method (standard compression)",
			method:   8,
			expected: "DEFLATE",
		},
		{
			name:     "unknown method 1",
			method:   1,
			expected: "0x1",
		},
		{
			name:     "unknown method 12",
			method:   12,
			expected: "0xC",
		},
		{
			name:     "unknown method 255",
			method:   255,
			expected: "0xFF",
		},
		{
			name:     "BZIP2 method (14)",
			method:   14,
			expected: "0xE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := methodToString(tt.method)
			if got != tt.expected {
				t.Errorf("methodToString(%d) = %v, want %v", tt.method, got, tt.expected)
			}
		})
	}
}

// Test Get File Argument Value verifies the parsing of command-line arguments
// to validate that only valid .zip files are accepted
func TestGetFileArgumentValue(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantFile  string
		wantError bool
		errorMsg  string
	}{
		{
			name:      "zip valid file",
			args:      []string{"program", "test.zip"},
			wantFile:  "test.zip",
			wantError: false,
		},
		{
			name:      "zip file with path",
			args:      []string{"program", "folder/test.zip"},
			wantFile:  "folder/test.zip",
			wantError: false,
		},
		{
			name:      "no zip file provided",
			args:      []string{"program"},
			wantFile:  "",
			wantError: true,
			errorMsg:  "no zip file provided",
		},
		{
			name:      "too many arguments",
			args:      []string{"program", "file1.zip", "file2.zip"},
			wantFile:  "",
			wantError: true,
			errorMsg:  "i don't know what to do with so many arguments",
		},
		{
			name:      "invalid zip file name",
			args:      []string{"program", "test.txt"},
			wantFile:  "",
			wantError: true,
			errorMsg:  "invalid zip file name",
		},
		{
			name:      "empty zip file name",
			args:      []string{"program", ""},
			wantFile:  "",
			wantError: true,
			errorMsg:  "invalid zip file name",
		},
		{
			name:      "only .zip extension",
			args:      []string{"program", ".zip"},
			wantFile:  ".zip",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()

			os.Args = tt.args

			got, err := getFileArgumentValue()

			if tt.wantError {
				if err == nil {
					t.Errorf("getFileArgumentValue() error = nil, wantError %v", tt.wantError)
					return
				}
				if err.Error() != tt.errorMsg {
					t.Errorf("getFileArgumentValue() error = %v, want %v", err.Error(), tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("getFileArgumentValue() unexpected error = %v", err)
					return
				}
				if got != tt.wantFile {
					t.Errorf("getFileArgumentValue() = %v, want %v", got, tt.wantFile)
				}
			}
		})
	}
}

// TestGetExecutionFolder checks that the execution directory can be obtained
func TestGetExecutionFolder(t *testing.T) {
	folder, err := getExecutionFolder()

	if err != nil {
		t.Errorf("getExecutionFolder() unexpected error = %v", err)
	}

	if folder == "" {
		t.Error("getExecutionFolder() returned empty string")
	}

	info, err := os.Stat(folder)
	if err != nil {
		t.Errorf("getExecutionFolder() returned non-existent directory: %v", err)
	}

	if !info.IsDir() {
		t.Error("getExecutionFolder() returned path is not a directory")
	}
}

// TestOpenZipFileErrors checks the error handling when opening ZIP files
func TestOpenZipFileErrors(t *testing.T) {
	t.Run("archivo no existente", func(t *testing.T) {
		_, err := openZipFile("/path/to/nonexistent/file.zip")
		if err == nil {
			t.Error("openZipFile() expected error for non-existent file, got nil")
		}
	})

	t.Run("archivo no es ZIP", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "notazip*.txt")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tmpFile.Name())

		tmpFile.WriteString("This is not a zip file")
		tmpFile.Close()

		_, err = openZipFile(tmpFile.Name())
		if err == nil {
			t.Error("openZipFile() expected error for non-zip file, got nil")
		}
	})
}

// TestOpenZipFileSuccess checks that valid ZIP files can be read
// Note: This test requires a test zip file in testdata/
func TestOpenZipFileSuccess(t *testing.T) {
	testZipPath := "testdata/test.zip"

	if _, err := os.Stat(testZipPath); os.IsNotExist(err) {
		t.Skip("Skipping test: testdata/test.zip not found. Create a test zip file to run this test.")
	}

	content, err := openZipFile(testZipPath)
	if err != nil {
		t.Fatalf("openZipFile() unexpected error = %v", err)
	}

	if content == nil {
		t.Error("openZipFile() returned nil content")
	}

	if len(content) == 0 {
		t.Error("openZipFile() returned empty content")
	}
}

// TestGetFileToExtractIntegration checks the integration flow of getting a file to extract
// Note: This test requires specific configuration and can be skipped in CI
func TestGetFileToExtractIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(oldDir)

	t.Run("archivo no encontrado en directorio actual", func(t *testing.T) {
		os.Args = []string{"program", "nonexistent.zip"}

		_, _, _, err := GetFileToExtract()
		if err == nil {
			t.Error("GetFileToExtract() expected error for non-existent file, got nil")
		}
	})

	t.Run("argumentos inválidos", func(t *testing.T) {
		os.Args = []string{"program"}

		_, _, _, err := GetFileToExtract()
		if err == nil {
			t.Error("GetFileToExtract() expected error for missing arguments, got nil")
		}
	})
}

// BenchmarkMethodToString measures the performance of method code to string conversion
func BenchmarkMethodToString(b *testing.B) {
	methods := []uint16{0, 8, 14, 255}

	for i := 0; i < b.N; i++ {
		for _, m := range methods {
			methodToString(m)
		}
	}
}

// BenchmarkGetFileArgumentValue measures the performance of argument parsing
func BenchmarkGetFileArgumentValue(b *testing.B) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"program", "test.zip"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		getFileArgumentValue()
	}
}
