package phy

import (
	"bytes"
	"encoding/binary"
	"io"
	"unsafe"
)

type Reader struct {
	stream io.Reader
	buf    []byte
}

func (reader *Reader) Read() (*Phy, error) {
	err := reader.getByteBuffer()
	if err != nil {
		return nil, err
	}

	// Read header
	header, err := reader.readHeader()
	if err != nil {
		return nil, err
	}

	offset := int32(0)

	//bodyparts
	offset += int32(unsafe.Sizeof(header))
	compacts, legacys, err := reader.readSolids(offset, header.SolidCount)
	if err != nil {
		return nil, err
	}

	return &Phy{
		Header:          header,
		CompactSurfaces: compacts,
		LegacySurfaces:  legacys,
	}, nil
}

// Reads phy header information
func (reader *Reader) readHeader() (header, error) {
	header := header{}
	headerSize := unsafe.Sizeof(header)

	err := binary.Read(bytes.NewBuffer(reader.buf[:headerSize]), binary.LittleEndian, &header)

	return header, err
}

// Read compact and legacy entries
func (reader *Reader) readSolids(offset int32, num int32) ([]compactSurfaceHeader, []legacySurfaceHeader, error) {
	compacts := make([]compactSurfaceHeader, num)
	legacys := make([]legacySurfaceHeader, num)
	compactSize := int32(unsafe.Sizeof(compactSurfaceHeader{}))
	legacySize := int32(unsafe.Sizeof(legacySurfaceHeader{}))

	for i := int32(0); i < num; i++ {
		//compact
		err := binary.Read(bytes.NewBuffer(reader.buf[offset:offset+compactSize]), binary.LittleEndian, &compacts[i])
		if err != nil {
			return compacts, legacys, err
		}

		offset += compactSize
		//legacy
		err = binary.Read(bytes.NewBuffer(reader.buf[offset:offset+legacySize]), binary.LittleEndian, &legacys[i])
		if err != nil {
			return compacts, legacys, err
		}
		offset += legacySize
	}

	return compacts, legacys, nil
}

// Read stream to []byte buffer
func (reader *Reader) getByteBuffer() error {
	buf := bytes.Buffer{}
	_, err := buf.ReadFrom(reader.stream)
	if err == nil {
		reader.buf = buf.Bytes()
	}

	return err
}
