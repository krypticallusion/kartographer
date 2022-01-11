package entities

import (
	"io"
	"log"
)

const FourKiB = 4096

// The Region Data structure formatted.
type Region struct {
	Locations  LocationTable
	Timestamps LastModifiedTimestampTable
}

// LocationTable gives us the location of chunks in the region file
type LocationTable struct {
	Table                   // is the raw table
	Entries []LocationEntry // Entries of locations
}

// LastModifiedTimestampTable gives us the last updated timestamp of the given chunk.
// LocationTable's index corresponds to the Timestamp index.
// i.e. LocationTable.Entries[0] corresponds to Timestamps.Entries[0].
type LastModifiedTimestampTable struct {
	Table   // is the raw table
	Entries []TimestampEntry
}

// ToEntries takes the raw table and converts it to a slice of entries.
// raw Table is a 1024 stack of entries.
func (l *LocationTable) ToEntries() error {
	entries, err := l.Table.Subdivide()
	if err != nil {
		return err
	}

	for _, e := range entries {
		l.Entries = append(l.Entries, LocationEntry{e})
	}

	return nil
}

// ToEntries takes the raw table and converts it to a slice of entries.
// raw Table is a 1024 stack of entries.
func (l *LastModifiedTimestampTable) ToEntries() error {
	entries, err := l.Table.Subdivide()
	if err != nil {
		return err
	}

	for _, e := range entries {
		l.Entries = append(l.Entries, TimestampEntry{e})
	}

	return nil
}

func NewRegion() *Region {
	return &Region{
		Locations:  LocationTable{Table: CreateInflatedTable(FourKiB)},
		Timestamps: LastModifiedTimestampTable{Table: CreateInflatedTable(FourKiB)},
	}
}

func LoadNewRegion(r io.ReadSeeker) (*Region, error) {
	region, err := NewRegion().Load(r)
	if err != nil {
		return nil, err
	}

	return region, nil
}

func (region *Region) Load(r io.ReadSeeker) (*Region, error) {
	region.Locations.Table.Inflate(FourKiB)
	region.Timestamps.Table.Inflate(FourKiB)

	if err := region.Locations.Read(r); err != nil {
		log.Println(err)
	}

	if err := region.Timestamps.Read(r); err != nil {
		log.Println(err)
	}

	_ = region.Locations.ToEntries()
	_ = region.Timestamps.ToEntries()

	return region, nil
}
