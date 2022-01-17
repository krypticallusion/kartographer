package utils

import (
	"errors"
	"fmt"
	"github.com/anthonynsimon/bild/adjust"
	"github.com/anthonynsimon/bild/transform"
	"image"
	"image/draw"
	"log"
	"os"
)

var RenderBlockErr = func(err error) error {
	return errors.New("RenderBlock: " + err.Error())
}

func RenderFullBlock(topPath string, sidePath string) (image.Image, error) {
	if topPath == "" || sidePath == "" {
		return nil, RenderBlockErr(fmt.Errorf("required topPath and sidePath textures"))
	}

	top, err := os.Open(topPath)
	if err != nil {
		log.Fatalf("RenderBlock: %s", err)
	}

	defer top.Close()

	side, err := os.Open(sidePath)
	if err != nil {
		return nil, RenderBlockErr(err)
	}

	defer side.Close()

	canvas := image.NewRGBA(image.Rectangle{Min: image.Point{}, Max: image.Point{X: 32, Y: 35}})

	topImage, _, err := image.Decode(top)
	if err != nil {
		return nil, RenderBlockErr(err)
	}

	sideImage, _, err := image.Decode(side)
	if err != nil {
		return nil, RenderBlockErr(err)
	}

	transformTop := transformTextures(topImage, transformOpts{
		rotationAngle: -45,
		w:             2.1,
	})

	transformLeft := adjust.Brightness(transformTextures(sideImage, transformOpts{
		shearVAngle: -30,
		h:           1.2,
	}), -0.2)

	transformRight := adjust.Brightness(transform.FlipH(transformLeft), -0.4)

	draw.Draw(canvas, transformTop.Bounds().Add(image.Point{X: 1, Y: 0}), transformTop, image.Point{}, draw.Over)
	draw.Draw(canvas, transformLeft.Bounds().Add(image.Point{X: 0, Y: 7}), transformLeft, image.Point{}, draw.Over)
	draw.Draw(canvas, transformRight.Bounds().Add(image.Point{X: 16, Y: 7}), transformRight, image.Point{}, draw.Over)

	return canvas, nil
}

type transformOpts struct {
	shearHAngle   float64
	shearVAngle   float64
	rotationAngle float64
	w             float64
	h             float64
}

// Resize vertically by a factor of h
// Resize horizontally by a factor of w
func transformTextures(img image.Image, opts transformOpts) image.Image {
	if opts.w == 0 {
		opts.w = 1
	}

	if opts.h == 0 {
		opts.h = 1
	}

	height, width := int(float64(img.Bounds().Dy())*opts.h), int(float64(img.Bounds().Dx())*opts.w)

	rotate := transform.Rotate(img, opts.rotationAngle, &transform.RotationOptions{ResizeBounds: true})
	resize := transform.Resize(rotate, width, height, transform.Lanczos)

	// Shear only if you need to
	// A shearAngle of 0 also distorts the image
	shear := resize

	if opts.shearHAngle != 0 {
		shear = transform.ShearH(resize, opts.shearHAngle)
	}

	if opts.shearVAngle != 0 {
		shear = transform.ShearV(shear, opts.shearVAngle)
	}

	return shear
}
