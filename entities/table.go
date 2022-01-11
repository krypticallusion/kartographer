package entities

import (
	"encoding/binary"
	"fmt"
	"io"
)

// Table is the structure type for the two tables that exist in a region file
type Table []byte

func (l *Table) Inflate(size int) {
	*l = make([]byte, size)
}

func (l *Table) Read(reader io.ReadSeeker) error {
	err := binary.Read(reader, binary.BigEndian, l)
	if err != nil {
		return err
	}

	return nil
}

func (l *Table) ReadWithOffset(reader io.ReadSeeker, offset int) error {
	_, err := reader.Seek(int64(offset), 0)
	if err != nil {
		return err
	}

	return l.Read(reader)
}

func CreateInflatedTable(size int) (t Table) {
	return make(Table, size)
}

// Subdivide the Table to Entries.
// There are going to be 1024 Entries in each Table.
// Each Entry is 4 bytes long.
func (l Table) Subdivide() (entries []Entry, err error) {
	if l == nil {
		return nil, fmt.Errorf("table is nil")
	}

	for i := 0; i < 1024; i++ {
		entries = append(entries, Entry(l[i*4:(i+1)*4]))
	}

	return
}
