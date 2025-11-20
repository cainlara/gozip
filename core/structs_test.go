package core

import "testing"

// TestNewZippedFile verifica que el constructor NewZippedFile
// inicialice correctamente todos los campos de la estructura ZippedFile
func TestNewZippedFile(t *testing.T) {
	tests := []struct {
		name       string
		fileName   string
		dir        bool
		size       uint64
		compressed uint64
		method     string
		modified   string
		crc        uint32
	}{
		{
			name:       "archivo regular con datos completos",
			fileName:   "test.txt",
			dir:        false,
			size:       1024,
			compressed: 512,
			method:     "DEFLATE",
			modified:   "2024-01-15T10:30:00Z",
			crc:        12345678,
		},
		{
			name:       "directorio",
			fileName:   "folder/",
			dir:        true,
			size:       0,
			compressed: 0,
			method:     "STORE",
			modified:   "2024-01-15T10:30:00Z",
			crc:        0,
		},
		{
			name:       "archivo sin comprimir",
			fileName:   "image.png",
			dir:        false,
			size:       2048,
			compressed: 2048,
			method:     "STORE",
			modified:   "2024-01-15T10:30:00Z",
			crc:        87654321,
		},
		{
			name:       "archivo vacío",
			fileName:   "empty.txt",
			dir:        false,
			size:       0,
			compressed: 0,
			method:     "STORE",
			modified:   "-",
			crc:        0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zf := NewZippedFile(tt.fileName, tt.dir, tt.size, tt.compressed, tt.method, tt.modified, tt.crc)

			if got := zf.GetName(); got != tt.fileName {
				t.Errorf("GetName() = %v, want %v", got, tt.fileName)
			}
			if got := zf.IsDir(); got != tt.dir {
				t.Errorf("IsDir() = %v, want %v", got, tt.dir)
			}
			if got := zf.GetSize(); got != tt.size {
				t.Errorf("GetSize() = %v, want %v", got, tt.size)
			}
			if got := zf.GetCompressedSize(); got != tt.compressed {
				t.Errorf("GetCompressedSize() = %v, want %v", got, tt.compressed)
			}
			if got := zf.GetMethod(); got != tt.method {
				t.Errorf("GetMethod() = %v, want %v", got, tt.method)
			}
			if got := zf.GetModifiedDate(); got != tt.modified {
				t.Errorf("GetModifiedDate() = %v, want %v", got, tt.modified)
			}
			if got := zf.GetCrc(); got != tt.crc {
				t.Errorf("GetCrc() = %v, want %v", got, tt.crc)
			}
		})
	}
}

// TestZippedFileGetters verifica individualmente cada método getter
// para asegurar que devuelven los valores correctos
func TestZippedFileGetters(t *testing.T) {
	zf := NewZippedFile("test.zip", false, 1024, 512, "DEFLATE", "2024-01-15T10:30:00Z", 12345678)

	t.Run("GetName", func(t *testing.T) {
		if got := zf.GetName(); got != "test.zip" {
			t.Errorf("GetName() = %v, want %v", got, "test.zip")
		}
	})

	t.Run("IsDir", func(t *testing.T) {
		if got := zf.IsDir(); got != false {
			t.Errorf("IsDir() = %v, want %v", got, false)
		}
	})

	t.Run("GetSize", func(t *testing.T) {
		if got := zf.GetSize(); got != 1024 {
			t.Errorf("GetSize() = %v, want %v", got, 1024)
		}
	})

	t.Run("GetCompressedSize", func(t *testing.T) {
		if got := zf.GetCompressedSize(); got != 512 {
			t.Errorf("GetCompressedSize() = %v, want %v", got, 512)
		}
	})

	t.Run("GetMethod", func(t *testing.T) {
		if got := zf.GetMethod(); got != "DEFLATE" {
			t.Errorf("GetMethod() = %v, want %v", got, "DEFLATE")
		}
	})

	t.Run("GetModifiedDate", func(t *testing.T) {
		if got := zf.GetModifiedDate(); got != "2024-01-15T10:30:00Z" {
			t.Errorf("GetModifiedDate() = %v, want %v", got, "2024-01-15T10:30:00Z")
		}
	})

	t.Run("GetCrc", func(t *testing.T) {
		if got := zf.GetCrc(); got != 12345678 {
			t.Errorf("GetCrc() = %v, want %v", got, 12345678)
		}
	})
}

// TestZippedFileDirectory verifica el comportamiento específico
// cuando se trabaja con directorios en archivos ZIP
func TestZippedFileDirectory(t *testing.T) {
	dir := NewZippedFile("my-folder/", true, 0, 0, "STORE", "2024-01-15T10:30:00Z", 0)

	if !dir.IsDir() {
		t.Error("Expected IsDir() to return true for directory")
	}

	if dir.GetSize() != 0 {
		t.Errorf("Expected directory size to be 0, got %d", dir.GetSize())
	}

	if dir.GetCrc() != 0 {
		t.Errorf("Expected directory CRC to be 0, got %d", dir.GetCrc())
	}
}

// TestZippedFileEdgeCases verifica casos extremos y valores límite
func TestZippedFileEdgeCases(t *testing.T) {
	t.Run("archivo con nombre muy largo", func(t *testing.T) {
		longName := "very/long/path/with/many/nested/directories/and/a/very/long/filename.txt"
		zf := NewZippedFile(longName, false, 100, 50, "DEFLATE", "2024-01-15T10:30:00Z", 999)

		if got := zf.GetName(); got != longName {
			t.Errorf("GetName() = %v, want %v", got, longName)
		}
	})

	t.Run("archivo con tamaño máximo uint64", func(t *testing.T) {
		maxSize := uint64(18446744073709551615) // max uint64
		zf := NewZippedFile("huge.bin", false, maxSize, maxSize/2, "DEFLATE", "2024-01-15T10:30:00Z", 0)

		if got := zf.GetSize(); got != maxSize {
			t.Errorf("GetSize() = %v, want %v", got, maxSize)
		}
	})

	t.Run("archivo con CRC máximo", func(t *testing.T) {
		maxCrc := uint32(4294967295) // max uint32
		zf := NewZippedFile("test.bin", false, 100, 50, "DEFLATE", "2024-01-15T10:30:00Z", maxCrc)

		if got := zf.GetCrc(); got != maxCrc {
			t.Errorf("GetCrc() = %v, want %v", got, maxCrc)
		}
	})

	t.Run("archivo con nombre vacío", func(t *testing.T) {
		zf := NewZippedFile("", false, 0, 0, "STORE", "-", 0)

		if got := zf.GetName(); got != "" {
			t.Errorf("GetName() = %v, want empty string", got)
		}
	})

	t.Run("método de compresión desconocido", func(t *testing.T) {
		zf := NewZippedFile("test.bin", false, 100, 50, "0x12", "2024-01-15T10:30:00Z", 123)

		if got := zf.GetMethod(); got != "0x12" {
			t.Errorf("GetMethod() = %v, want %v", got, "0x12")
		}
	})
}
