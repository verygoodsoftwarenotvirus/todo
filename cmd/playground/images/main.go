package main

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"
)

func buildArbitraryImage(width, height int) image.Image {
	img := image.NewRGBA(image.Rectangle{Min: image.Point{}, Max: image.Point{X: width, Y: height}})

	// Set color for each pixel.
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			img.Set(x, y, color.RGBA{R: uint8(x % math.MaxUint8), G: uint8(y % math.MaxUint8), B: uint8(x + y%math.MaxUint8), A: math.MaxUint8})
		}
	}

	return img
}

func main() {
	width := 256
	height := 256

	img := image.NewRGBA(image.Rectangle{Min: image.Point{}, Max: image.Point{X: width, Y: height}})

	// Set color for each pixel.
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			img.Set(x, y, color.RGBA{R: uint8(x % 255), G: uint8(y % 255), B: uint8(x + y%255), A: 255})
		}
	}

	// Encode as PNG.
	f, _ := os.Create("artifacts/test_image.png")

	if err := png.Encode(f, img); err != nil {
		log.Fatal(err)
	}
}
