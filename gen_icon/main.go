package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"

	qrcode "github.com/skip2/go-qrcode"
)

func main() {
	size := 1024
	padding := 80
	cornerRadius := 180.0

	// Generate QR code bitmap
	qr, err := qrcode.New("https://qrcode.app", qrcode.High)
	if err != nil {
		panic(err)
	}
	qr.DisableBorder = true
	qrImg := qr.Image(size - padding*2)
	qrSize := qrImg.Bounds().Max.X

	// Background color: deep indigo
	bg := color.NRGBA{R: 30, G: 30, B: 50, A: 255}
	// QR module color: white
	fg := color.NRGBA{R: 255, G: 255, B: 255, A: 255}

	// Create output image
	out := image.NewNRGBA(image.Rect(0, 0, size, size))

	// Fill with background using rounded rect mask
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			if inRoundedRect(x, y, size, cornerRadius) {
				out.Set(x, y, bg)
			}
		}
	}

	// Draw QR code pixels
	offsetX := (size - qrSize) / 2
	offsetY := (size - qrSize) / 2

	for y := 0; y < qrSize; y++ {
		for x := 0; x < qrSize; x++ {
			src := qrImg.At(x, y)
			r, g, b, _ := src.RGBA()
			// go-qrcode: dark modules are black (0,0,0)
			if r == 0 && g == 0 && b == 0 {
				px := offsetX + x
				py := offsetY + y
				if inRoundedRect(px, py, size, cornerRadius) {
					out.Set(px, py, fg)
				}
			}
		}
	}

	// Add subtle corner squares highlight (finder patterns already in QR)
	// Draw a thin border inside the icon for polish
	borderWidth := 6
	borderColor := color.NRGBA{R: 100, G: 100, B: 180, A: 120}
	drawRoundedBorder(out, size, cornerRadius, borderWidth, borderColor)

	_ = draw.Op(0) // avoid unused import

	f, err := os.Create("build/appicon.png")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err := png.Encode(f, out); err != nil {
		panic(err)
	}
}

func inRoundedRect(x, y, size int, r float64) bool {
	fx, fy := float64(x)+0.5, float64(y)+0.5
	s := float64(size)
	if fx < r && fy < r {
		return dist(fx, fy, r, r) <= r
	}
	if fx > s-r && fy < r {
		return dist(fx, fy, s-r, r) <= r
	}
	if fx < r && fy > s-r {
		return dist(fx, fy, r, s-r) <= r
	}
	if fx > s-r && fy > s-r {
		return dist(fx, fy, s-r, s-r) <= r
	}
	return true
}

func dist(x1, y1, x2, y2 float64) float64 {
	dx, dy := x1-x2, y1-y2
	return math.Sqrt(dx*dx + dy*dy)
}

func drawRoundedBorder(img *image.NRGBA, size int, r float64, bw int, c color.NRGBA) {
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			in := inRoundedRect(x, y, size, r)
			inInner := inRoundedRect(x-bw, y-bw, size-bw*2, r*float64(size-bw*2)/float64(size))
			if in && !inInner {
				// blend
				dst := img.NRGBAAt(x, y)
				img.Set(x, y, blend(dst, c))
			}
		}
	}
}

func blend(dst, src color.NRGBA) color.NRGBA {
	a := float64(src.A) / 255.0
	return color.NRGBA{
		R: uint8(float64(dst.R)*(1-a) + float64(src.R)*a),
		G: uint8(float64(dst.G)*(1-a) + float64(src.G)*a),
		B: uint8(float64(dst.B)*(1-a) + float64(src.B)*a),
		A: 255,
	}
}
