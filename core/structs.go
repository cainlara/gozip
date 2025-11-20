package core

type ZippedFile struct {
	fileName   string
	dir        bool
	size       uint64
	compressed uint64
	method     string
	modified   string
	crc        uint32
}

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

func (zf ZippedFile) GetName() string {
	return zf.fileName
}

func (zf ZippedFile) IsDir() bool {
	return zf.dir
}

func (zf ZippedFile) GetSize() uint64 {
	return zf.size
}

func (zf ZippedFile) GetCompressedSize() uint64 {
	return zf.compressed
}

func (zf ZippedFile) GetMethod() string {
	return zf.method
}

func (zf ZippedFile) GetModifiedDate() string {
	return zf.modified
}

func (zf ZippedFile) GetCrc() uint32 {
	return zf.crc
}
