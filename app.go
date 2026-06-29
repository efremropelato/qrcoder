package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	stdraw "image/draw"
	_ "image/gif"
	"image/jpeg"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strings"

	qrcode "github.com/skip2/go-qrcode"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	xdraw "golang.org/x/image/draw"
)

type App struct {
	ctx context.Context
}

// QRParams holds all parameters for QR code generation.
type QRParams struct {
	URL           string `json:"url"`
	Size          int    `json:"size"`
	FgColor       string `json:"fgColor"`
	BgColor       string `json:"bgColor"`
	LogoPath      string `json:"logoPath"`
	ErrorLevel    string `json:"errorLevel"`    // L, M, Q, H (forced to H when logo present)
	Format        string `json:"format"`        // png, jpeg
	DisableBorder bool   `json:"disableBorder"` // remove quiet-zone border
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// GenerateQRCode builds the QR image and returns it as a base64-encoded PNG string
// for live preview in the frontend.
func (a *App) GenerateQRCode(params QRParams) (string, error) {
	img, err := buildQRImage(params)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", fmt.Errorf("encode png: %w", err)
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

// SaveQRCode opens a native save dialog and writes the QR image to disk.
// Returns the final file path, or empty string if the user cancelled.
func (a *App) SaveQRCode(params QRParams) (string, error) {
	format := strings.ToLower(params.Format)
	if format != "jpeg" {
		format = "png"
	}

	var filters []runtime.FileFilter
	var defaultName string
	if format == "jpeg" {
		filters = []runtime.FileFilter{
			{DisplayName: "JPEG Image (*.jpg)", Pattern: "*.jpg;*.jpeg"},
			{DisplayName: "PNG Image (*.png)", Pattern: "*.png"},
		}
		defaultName = "qrcode.jpg"
	} else {
		filters = []runtime.FileFilter{
			{DisplayName: "PNG Image (*.png)", Pattern: "*.png"},
			{DisplayName: "JPEG Image (*.jpg)", Pattern: "*.jpg;*.jpeg"},
		}
		defaultName = "qrcode.png"
	}

	savePath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		DefaultFilename: defaultName,
		Filters:         filters,
	})
	if err != nil || savePath == "" {
		return "", err
	}

	img, err := buildQRImage(params)
	if err != nil {
		return "", err
	}

	f, err := os.Create(savePath)
	if err != nil {
		return "", fmt.Errorf("crea file: %w", err)
	}
	defer f.Close()

	switch strings.ToLower(filepath.Ext(savePath)) {
	case ".jpg", ".jpeg":
		if err := jpeg.Encode(f, img, &jpeg.Options{Quality: 95}); err != nil {
			return "", fmt.Errorf("encode jpeg: %w", err)
		}
	default:
		if err := png.Encode(f, img); err != nil {
			return "", fmt.Errorf("encode png: %w", err)
		}
	}

	return savePath, nil
}

// SelectLogoFile opens a native file-picker filtered to common image formats.
func (a *App) SelectLogoFile() (string, error) {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Seleziona Logo",
		Filters: []runtime.FileFilter{
			{DisplayName: "Immagini (*.png;*.jpg;*.gif)", Pattern: "*.png;*.jpg;*.jpeg;*.gif"},
		},
	})
	if err != nil {
		return "", err
	}
	return path, nil
}

func buildQRImage(params QRParams) (image.Image, error) {
	if strings.TrimSpace(params.URL) == "" {
		return nil, fmt.Errorf("URL obbligatorio")
	}
	if params.Size < 64 {
		params.Size = 64
	}
	if params.Size > 4096 {
		params.Size = 4096
	}

	fg, err := parseHexColor(params.FgColor)
	if err != nil {
		fg = color.Black
	}
	bg, err := parseHexColor(params.BgColor)
	if err != nil {
		bg = color.White
	}

	level := parseErrorLevel(params.ErrorLevel, params.LogoPath)

	qr, err := qrcode.New(params.URL, level)
	if err != nil {
		return nil, fmt.Errorf("genera qr: %w", err)
	}
	qr.ForegroundColor = fg
	qr.BackgroundColor = bg
	qr.DisableBorder = params.DisableBorder

	qrImg := qr.Image(params.Size)

	if params.LogoPath == "" {
		return qrImg, nil
	}

	logoSrc, err := loadImage(params.LogoPath)
	if err != nil {
		return nil, fmt.Errorf("carica logo: %w", err)
	}

	return overlayLogo(qrImg, logoSrc, params.Size), nil
}

// parseErrorLevel returns the recovery level for the QR code.
// When a logo is present, H (Highest) is always used because the logo
// covers ~25% of the center and requires maximum error correction.
func parseErrorLevel(level, logoPath string) qrcode.RecoveryLevel {
	if logoPath != "" {
		return qrcode.Highest
	}
	switch strings.ToUpper(level) {
	case "L":
		return qrcode.Low
	case "M":
		return qrcode.Medium
	case "Q":
		return qrcode.High
	default:
		return qrcode.Highest
	}
}

func overlayLogo(base image.Image, logo image.Image, size int) image.Image {
	logoSize := int(math.Round(float64(size) * 0.25))

	dst := image.NewRGBA(base.Bounds())
	stdraw.Draw(dst, dst.Bounds(), base, image.Point{}, stdraw.Src)

	scaledLogo := image.NewRGBA(image.Rect(0, 0, logoSize, logoSize))
	xdraw.CatmullRom.Scale(scaledLogo, scaledLogo.Bounds(), logo, logo.Bounds(), xdraw.Over, nil)

	offset := image.Point{
		X: (size - logoSize) / 2,
		Y: (size - logoSize) / 2,
	}
	logoRect := image.Rectangle{Min: offset, Max: offset.Add(image.Point{X: logoSize, Y: logoSize})}
	stdraw.Draw(dst, logoRect, scaledLogo, image.Point{}, stdraw.Over)

	return dst
}

// loadImage decodes any registered image format (PNG, JPEG, GIF) from disk.
// Using image.Decode auto-detects the format from magic bytes instead of extension.
func loadImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("decodifica immagine: %w", err)
	}
	return img, nil
}

func parseHexColor(s string) (color.Color, error) {
	s = strings.TrimPrefix(s, "#")
	switch len(s) {
	case 6:
		var r, g, b uint8
		if _, err := fmt.Sscanf(s, "%02x%02x%02x", &r, &g, &b); err != nil {
			return nil, err
		}
		return color.RGBA{R: r, G: g, B: b, A: 255}, nil
	case 8:
		var r, g, b, a uint8
		if _, err := fmt.Sscanf(s, "%02x%02x%02x%02x", &r, &g, &b, &a); err != nil {
			return nil, err
		}
		return color.RGBA{R: r, G: g, B: b, A: a}, nil
	default:
		return nil, fmt.Errorf("colore non valido: #%s", s)
	}
}
