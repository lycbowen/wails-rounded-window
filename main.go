package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
)

//go:embed all:frontend/dist
var assets embed.FS

const appTitle = "wails-rounded-window"

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:         appTitle,
		Width:         1024,
		Height:        768,
		MinWidth:      720,
		MinHeight:     480,
		Frameless:     true,
		DisableResize: false,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: options.NewRGBA(0, 0, 0, 0),
		Mac: &mac.Options{
			WindowIsTranslucent:  true,
			WebviewIsTransparent: true,
			TitleBar: &mac.TitleBar{
				TitlebarAppearsTransparent: true,
				HideTitleBar:               true,
				HideTitle:                  true,
				FullSizeContent:            true,
			},
		},
		CSSDragProperty: "--wails-draggable",
		CSSDragValue:    "drag",
		OnStartup:       app.startup,
		OnDomReady:      app.domReady,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
