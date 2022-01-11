package entities

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
)

type Entry []byte

type LocationEntry struct {
	Entry
}

type TimestampEntry struct {
	Entry
}

// GetOffset converts the first 3 bytes of an entry are the offset bytes of the chunk
func (e LocationEntry) GetOffset() int {
	return int(uint(e.Entry[2]) | uint(e.Entry[1])<<8 | uint(e.Entry[0])<<16)
}

func (e LocationEntry) GetSize() int {
	return int(e.Entry[3])
}

func (e LocationEntry) GetActualRangeOfEntry() (int, int) {
	return e.GetOffset() * FourKiB, e.GetSize() * FourKiB
}

// GetChunk retrieves the actual chunk that exists in the region file
// Trimmed according to retrieved entry ranges.
func (e LocationEntry) GetChunk(r io.ReadSeeker) (c Chunk, err error) {
	start, size := e.GetActualRangeOfEntry()
	if start == 0 && size == 0 {
		return Chunk{}, fmt.Errorf("GetChunk: chunk not generated, yet")
	}

	chunk := make([]byte, size)

	_, err = r.Seek(int64(start), 0)
	if err != nil {
		return Chunk{}, err
	}

	_, err = r.Read(chunk)
	if err != nil {
		return Chunk{}, err
	}

	c.raw = chunk

	if err := c.ExtractFromRaw(); err != nil {
		log.Println(err)
	}

	return
}

func (t TimestampEntry) GetTimestamp() (int, error) {
	return int(binary.BigEndian.Uint32(t.Entry)), nil
}
