// Package core provides the fundamental data structures for representing
// files within a ZIP archive.
package core

// ZippedFile represents a file or directory within a ZIP archive.
// It contains all metadata associated with the compressed file, including
// name, size, compression method, modification date, and CRC.
type ZippedFile struct {
	fileName   string
	dir        bool
	size       uint64
	compressed uint64
	method     string
	modified   string
	crc        uint32
}

// NewZippedFile creates a new ZippedFile instance with the provided parameters.
// This constructor initializes all fields of the structure with the given values.
//
// Parameters:
//   - fileName: name of the file or directory within the ZIP
//   - dir: true if it's a directory, false if it's a file
//   - size: uncompressed size in bytes
//   - compressed: compressed size in bytes
//   - method: compression method used (e.g., "STORE", "DEFLATE")
//   - modified: modification date in RFC3339 format
//   - crc: CRC32 value of the file
func NewZippedFile(fileName string, dir bool, size uint64, compressed uint64, method string, modified string, crc uint32) ZippedFile {
	return ZippedFile{
		fileName:   fileName,
		dir:        dir,
		size:       size,
		compressed: compressed,
		method:     method,
		modified:   modified,
		crc:        crc,
	}
}

// GetName returns the name of the file or directory within the ZIP.
func (zf ZippedFile) GetName() string {
	return zf.fileName
}

// IsDir returns true if the ZippedFile represents a directory, false if it's a file.
func (zf ZippedFile) IsDir() bool {
	return zf.dir
}

// GetSize returns the uncompressed size of the file in bytes.
// For directories, this value is typically 0.
func (zf ZippedFile) GetSize() uint64 {
	return zf.size
}

// GetCompressedSize returns the compressed size of the file in bytes.
// This value may equal the uncompressed size if the STORE method was used.
func (zf ZippedFile) GetCompressedSize() uint64 {
	return zf.compressed
}

// GetMethod returns the compression method used for the file.
// Common values are "STORE" (no compression) and "DEFLATE" (standard compression).
func (zf ZippedFile) GetMethod() string {
	return zf.method
}

// GetModifiedDate returns the modification date of the file in RFC3339 format.
// Returns "-" if the date is not available.
func (zf ZippedFile) GetModifiedDate() string {
	return zf.modified
}

// GetCrc returns the CRC32 value of the file.
// This value is used to verify the file's integrity.
func (zf ZippedFile) GetCrc() uint32 {
	return zf.crc
}
