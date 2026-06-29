package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:     "QR Code Generator",
		Width:     900,
		Height:    660,
		MinWidth:  720,
		MinHeight: 520,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 15, G: 17, B: 23, A: 255},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId: "com.ropelato.qrcodegenerator-v1",
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			Theme:                windows.Dark,
			CustomTheme: &windows.ThemeSettings{
				DarkModeTitleBar:   windows.RGB(15, 17, 23),
				DarkModeTitleText:  windows.RGB(226, 228, 240),
				DarkModeBorder:     windows.RGB(46, 51, 80),
				LightModeTitleBar:  windows.RGB(240, 240, 240),
				LightModeTitleText: windows.RGB(20, 20, 20),
				LightModeBorder:    windows.RGB(200, 200, 200),
			},
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
