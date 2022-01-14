package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/fs"
	"kartographer/entities"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const RegionDir = "/home/k/.minecraft/saves/test/region"

func main() {
	tt := time.Now()

	// This aren't the actual proportions
	// just a random big enough canvas for drawing my sample
	img := image.NewRGBA(image.Rectangle{Min: image.Point{X: -512, Y: -512}, Max: image.Point{X: 512, Y: 512}})

	// Draws a white background
	draw.Draw(img, img.Bounds(), &image.Uniform{C: color.White}, image.Point{}, draw.Src)

	var rwg sync.WaitGroup

	_ = filepath.WalkDir(RegionDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Println(err)
			return err
		}

		if d.IsDir() {
			return err
		}

		rwg.Add(1)
		go func(rwg *sync.WaitGroup, path string) {
			defer rwg.Done()

			f, err := os.Open(path)
			if err != nil {
				log.Fatal(err)
			}

			r, err := entities.LoadNewRegion(f)
			if err != nil {
				log.Println(err)
			}

			for _, entry := range r.Locations.Entries {
				chunk, err := entry.GetChunk(f)
				if err != nil {
					continue
				}

				chunkImage := chunk.DrawORT()

				draw.Draw(img, chunkImage.Bounds().Add(image.Point{X: chunk.NBT.XPos * 16, Y: chunk.NBT.ZPos * 16}), chunkImage, image.Point{}, draw.Over)
			}
		}(&rwg, path)
		return nil
	})

	rwg.Wait()

	ts := time.Now()

	ff, err := os.Create("rectangle.png")
	defer ff.Close()
	if err != nil {
		log.Println(err)
	}
	_ = png.Encode(ff, img)

	fmt.Printf("Time taken: %s\n", ts.Sub(tt))
}
