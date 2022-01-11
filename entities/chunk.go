package entities

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"encoding/binary"
	"github.com/Tnze/go-mc/nbt"
	"image"
	"io"
	"karte/textures"
	"karte/utils"
	"math"
)

type CompressionScheme int

const (
	NoCompression   CompressionScheme = 0
	GzipCompression CompressionScheme = 1
	ZlibCompression CompressionScheme = 2
)

type Chunk struct {
	raw               []byte // the raw chunk data
	rawNbt            []byte // the raw chunk nbt data, chunk header removed
	NBT               ChunkNBT
	Length            int
	CompressionScheme CompressionScheme
}

type ChunkNBT struct {
	DataVersion   int        `nbt:"DataVersion"`
	XPos          int        `nbt:"xPos"`
	ZPos          int        `nbt:"zPos"`
	Status        string     `nbt:"Status"`
	isLightOn     int32      `nbt:"isLightOn"`
	InhabitedTime int64      `nbt:"InhabitedTime"`
	LastUpdate    int64      `nbt:"LastUpdate"`
	Sections      []Section  `nbt:"sections"` // Sections are 16x16x16 areas. Typically, a sub-chunk of a chunk
	HeightMaps    HeightMap  `nbt:"Heightmaps"`
	Structures    Structures `nbt:"structures"`
}

type HeightMap struct {
	MotionBlocking         []int64 `nbt:"MOTION_BLOCKING"`
	WorldSurface           []int64 `nbt:"WORLD_SURFACE"`
	WorldSurfaceWg         []int64 `nbt:"WORLD_SURFACE_WG"`
	MotionBlockingNoLeaves []int64 `nbt:"MOTION_BLOCKING_NO_LEAVES"`
	OceanFloor             []int64 `nbt:"OCEAN_FLOOR"`
	OceanFloorWg           []int64 `nbt:"OCEAN_FLOOR_WG"`
}

type Biomes struct {
	Palette []string `nbt:"palette"`
}

type BlockStates struct {
	Data    []int64 `nbt:"data"`
	Palette []struct {
		Name       string            `nbt:"Name"`
		Properties map[string]string `nbt:"Properties"`
	} `nbt:"palette"`
}

// ProcessData Processes the bits and puts them all into one array
// Block at respective X,Y,Z coordinates (relative to the section)
// would then be able to be accessed by y*256 + z*16 + x
func (b BlockStates) ProcessData() []int {
	//bitsPerBlock := len(b.Data)*64/4096
	indices := make([]int, len(b.Data)*64/4)
	currentIndex := 0

	for _, v := range b.Data {
		bits := utils.IntToBits(v)

		for i := 60; i >= 0; i -= 4 {
			indices[currentIndex] = utils.BitsToInt(bits[i : i+4])
			currentIndex += 1
		}
	}

	return indices
}

type Section struct {
	Y           int         `nbt:"Y"`
	BlockLight  []byte      `nbt:"BlockLight"`
	SkyLight    []byte      `nbt:"SkyLight"`
	Biomes      Biomes      `nbt:"biomes"`
	BlockStates BlockStates `nbt:"block_states"`
}

type Structures struct {
	References struct {
		BastionRemnant []int64 `nbt:"bastion_remnant"`
		OceanRuin      []int64 `nbt:"ocean_ruin"`
	}
}

func (c Chunk) New(start int, size int, from io.ReadSeeker) error {
	raw := make([]byte, size)
	_, err := from.Seek(int64(start), 0)
	if err != nil {
		return err
	}

	_, err = from.Read(raw)
	if err != nil {
		return err
	}

	c.raw = raw
	return nil
}

func (c *Chunk) ExtractFromRaw() error {
	header := c.raw[:5]

	c.Length = int(binary.BigEndian.Uint32(header[:4]))

	switch int(uint(header[4])) {
	case 0:
		c.CompressionScheme = NoCompression
	case 1:
		c.CompressionScheme = GzipCompression
	case 2:
		c.CompressionScheme = ZlibCompression
	}

	c.rawNbt = c.raw[5:]
	_ = c.FromNBT()

	return nil
}

func (c *Chunk) FromNBT() error {
	var reader io.Reader
	var err error

	switch c.CompressionScheme {
	case NoCompression:
		reader = bytes.NewReader(c.rawNbt)
	case GzipCompression:
		reader, err = gzip.NewReader(bytes.NewReader(c.rawNbt))
	case ZlibCompression:
		reader, err = zlib.NewReader(bytes.NewReader(c.rawNbt))
	}

	if err != nil {
		return err
	}

	_, err = nbt.NewDecoder(reader).Decode(&c.NBT)
	if err != nil {
		return err
	}

	return nil
}

// DrawORT draws the chunk orthogonally (top-to-bottom).
//
// It draws directly on a canvas
//
// Each pixel is one block.
func (c Chunk) DrawORT() *image.RGBA {
	img := image.NewRGBA(image.Rectangle{Min: image.Point{X: 0, Y: 0}, Max: image.Point{X: 16, Y: 16}})

	// ChunkX -> X coord within Chunk
	// ChunkZ -> Z coord within Chunk
	// Abs X, Z = X+Chunk.X*16
	ChunkX := 0
	ChunkZ := 0

	for im, m := range c.NBT.HeightMaps.MotionBlocking {
		b := utils.IntToBits(m)

		// In the Heightmap, each entry is of 9bits
		// 37 longs exists in a HeightMap
		// so 9*37 = 259
		// since the a chunk has 256 blocks on a certain Y level
		// the remaining 3 entries are not needed
		// and are hence empty
		endAt := 0

		// endAt checks sets the ending Index at appropriate index
		if im == len(c.NBT.HeightMaps.MotionBlocking)-1 {
			// the last 3 entries are unused in the last long
			endAt = 28
		}

		// Traverse through all the entries in a bit
		// Since 9 entries and the entries are reversed in order
		// The First entry is from 55 to 63, and so on.
		// Slice operator excludes the last index
		for i := 55; i >= endAt; i -= 9 {
			absY := utils.BitsToInt(b[i:i+9]) - 65

			// Since a chunk is created of 16 Sections
			// We can use the absY to find where the absY resides
			// In a section, which then we can traverse in
			closestSection := int(math.Floor(float64(absY) / float64(16)))

			// After every 16 intervals, we move to a new Z coord
			// | x0 x1 x2 .. x15 | z0
			// | x0 x1 x2 .. x15 | z1...
			// | x0 x1 x2 .. x15 | z15
			if ChunkX == 16 {
				ChunkX = 0
				ChunkZ += 1
			}

			// +4 is due to the recent 1.18 update which brings the world below 0 (-64)
			// 252,253,254,255 are the first 4 sections in the data
			// We'll handle this later tbh
			// i.e if the highest block in certain x,z  is below 0
			s := c.NBT.Sections[closestSection+4]
			if len(s.BlockStates.Data) == 0 {
				continue
			}

			// We get the local Y coord within the section
			y := absY - closestSection*16

			bpb := len(s.BlockStates.Data) * 64 / 4096

			// Finding the usable bits that exists in one long
			// if bpb=4, it uses all of the long, since 64 is divisible by 4 : 64 usable
			// if bpb=5, the closest multiple answer is 60, 12*5 : 60 usable
			// if bpb=6, the closest multiple answer is 60 as well, 10*6
			usableBits := 64 - 64%bpb

			blocksInALong := usableBits / bpb

			// Find the long where the required x,y,z resides
			// We first find how many blocks we have to travel
			// i.e y*256+z*16
			// Then multiplying with bpb, we get how many bits we have to travel
			// Then dividing by usableBits, we get how many longs we have to travel
			long := s.BlockStates.Data[int(math.Floor(float64((y*256+ChunkZ*16)*bpb/usableBits)))]

			longBits := utils.IntToBits(long)

			// ChunkX%blocksInALong since blocks on a certain Z coord can lie on different longs
			// We need to find where the block is, with respect
			// to the long
			// eg: On bpb 5 : (x,z) : 13,0 lies on the second long (12)
			// Whereas on bpb 4: (x,z) : 13,0 lies on the first long itself, since there is space (16)
			index := utils.BitsToInt(longBits[64-(ChunkX%blocksInALong)*bpb-bpb : 64-(ChunkX%blocksInALong)*bpb])
			img.Set(ChunkX, ChunkZ, textures.ColorPalette[s.BlockStates.Palette[index].Name])
			ChunkX += 1

		}
	}

	return img
}
