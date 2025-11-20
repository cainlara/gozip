package util

import (
	"os"
	"testing"
)

// TestMethodToString verifica que la conversión de códigos de método
// a strings funcione correctamente para los métodos de compresión estándar
func TestMethodToString(t *testing.T) {
	tests := []struct {
		name     string
		method   uint16
		expected string
	}{
		{
			name:     "método STORE (sin compresión)",
			method:   0,
			expected: "STORE",
		},
		{
			name:     "método DEFLATE (compresión estándar)",
			method:   8,
			expected: "DEFLATE",
		},
		{
			name:     "método desconocido 1",
			method:   1,
			expected: "0x1",
		},
		{
			name:     "método desconocido 12",
			method:   12,
			expected: "0xC",
		},
		{
			name:     "método desconocido 255",
			method:   255,
			expected: "0xFF",
		},
		{
			name:     "método BZIP2 (14)",
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

// TestGetFileArgumentValue verifica el parsing de argumentos de línea de comandos
// para validar que solo se acepten archivos .zip válidos
func TestGetFileArgumentValue(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantFile  string
		wantError bool
		errorMsg  string
	}{
		{
			name:      "archivo zip válido",
			args:      []string{"program", "test.zip"},
			wantFile:  "test.zip",
			wantError: false,
		},
		{
			name:      "archivo zip con ruta",
			args:      []string{"program", "folder/test.zip"},
			wantFile:  "folder/test.zip",
			wantError: false,
		},
		{
			name:      "sin argumentos",
			args:      []string{"program"},
			wantFile:  "",
			wantError: true,
			errorMsg:  "no zip file provided",
		},
		{
			name:      "demasiados argumentos",
			args:      []string{"program", "file1.zip", "file2.zip"},
			wantFile:  "",
			wantError: true,
			errorMsg:  "i don't know what to do with so many arguments",
		},
		{
			name:      "extensión incorrecta",
			args:      []string{"program", "test.txt"},
			wantFile:  "",
			wantError: true,
			errorMsg:  "invalid zip file name",
		},
		{
			name:      "nombre vacío",
			args:      []string{"program", ""},
			wantFile:  "",
			wantError: true,
			errorMsg:  "invalid zip file name",
		},
		{
			name:      "solo extensión .zip",
			args:      []string{"program", ".zip"},
			wantFile:  ".zip",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Guardar los argumentos originales
			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()

			// Establecer los argumentos de prueba
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

// TestGetExecutionFolder verifica que se pueda obtener el directorio de ejecución
func TestGetExecutionFolder(t *testing.T) {
	folder, err := getExecutionFolder()

	if err != nil {
		t.Errorf("getExecutionFolder() unexpected error = %v", err)
	}

	if folder == "" {
		t.Error("getExecutionFolder() returned empty string")
	}

	// Verificar que el directorio existe
	info, err := os.Stat(folder)
	if err != nil {
		t.Errorf("getExecutionFolder() returned non-existent directory: %v", err)
	}

	if !info.IsDir() {
		t.Error("getExecutionFolder() returned path is not a directory")
	}
}

// TestOpenZipFileErrors verifica el manejo de errores al abrir archivos ZIP
func TestOpenZipFileErrors(t *testing.T) {
	t.Run("archivo no existente", func(t *testing.T) {
		_, err := openZipFile("/path/to/nonexistent/file.zip")
		if err == nil {
			t.Error("openZipFile() expected error for non-existent file, got nil")
		}
	})

	t.Run("archivo no es ZIP", func(t *testing.T) {
		// Crear un archivo temporal que no sea ZIP
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

// TestOpenZipFileSuccess verifica que se puedan leer archivos ZIP válidos
// Nota: Este test requiere un archivo ZIP de prueba en testdata/
func TestOpenZipFileSuccess(t *testing.T) {
	// Crear un archivo ZIP de prueba simple
	testZipPath := "testdata/test.zip"

	// Verificar si el archivo de prueba existe
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

	// Verificar que el contenido sea un slice válido
	if len(content) < 0 {
		t.Error("openZipFile() returned invalid content length")
	}
}

// TestGetFileToExtractIntegration es un test de integración que verifica
// el flujo completo de obtención de archivo para extraer
// Nota: Este test requiere configuración específica y puede ser skipped en CI
func TestGetFileToExtractIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Guardar los argumentos originales
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Guardar el directorio actual
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(oldDir)

	t.Run("archivo no encontrado en directorio actual", func(t *testing.T) {
		os.Args = []string{"program", "nonexistent.zip"}

		_, _, err := GetFileToExtract()
		if err == nil {
			t.Error("GetFileToExtract() expected error for non-existent file, got nil")
		}
	})

	t.Run("argumentos inválidos", func(t *testing.T) {
		os.Args = []string{"program"}

		_, _, err := GetFileToExtract()
		if err == nil {
			t.Error("GetFileToExtract() expected error for missing arguments, got nil")
		}
	})
}

// BenchmarkMethodToString mide el rendimiento de la conversión de métodos
func BenchmarkMethodToString(b *testing.B) {
	methods := []uint16{0, 8, 14, 255}

	for i := 0; i < b.N; i++ {
		for _, m := range methods {
			methodToString(m)
		}
	}
}

// BenchmarkGetFileArgumentValue mide el rendimiento del parsing de argumentos
func BenchmarkGetFileArgumentValue(b *testing.B) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"program", "test.zip"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		getFileArgumentValue()
	}
}
