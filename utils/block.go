package utils

import (
	"errors"
	"fmt"
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

	canvas := image.NewRGBA(image.Rectangle{Min: image.Point{}, Max: image.Point{X: 32, Y: 32}})

	topImage, _, err := image.Decode(top)
	if err != nil {
		return nil, RenderBlockErr(err)
	}

	sideImage, _, err := image.Decode(side)
	if err != nil {
		return nil, RenderBlockErr(err)
	}

	transformTop := transformTextures(topImage, transformOpts{
		shearHAngle:   26.5650512,
		rotationAngle: 30,
		h:             0.864,
	})

	transformLeft := transformTextures(sideImage, transformOpts{
		shearHAngle:   -26.5650512,
		rotationAngle: 30,
		h:             0.864,
	})

	transformRight := transform.FlipH(transformLeft)

	draw.Draw(canvas, transformTop.Bounds().Add(image.Point{X: 4, Y: -1}), transformTop, image.Point{}, draw.Over)
	draw.Draw(canvas, transformLeft.Bounds().Add(image.Point{X: -2, Y: 10}), transformLeft, image.Point{}, draw.Over)
	draw.Draw(canvas, transformRight.Bounds().Add(image.Point{X: 9, Y: 10}), transformRight, image.Point{}, draw.Over)

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

	resize := transform.Resize(img, width, height, transform.NearestNeighbor)
	shear := transform.ShearH(resize, opts.shearHAngle)
	shear = transform.ShearV(shear, opts.shearVAngle)
	rotate := transform.Rotate(shear, opts.rotationAngle, &transform.RotationOptions{ResizeBounds: true})

	return rotate
}
