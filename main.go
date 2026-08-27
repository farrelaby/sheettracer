package main

import (
	"embed"
	"log"
	"os"
	"path/filepath"
	"runtime"

	database "sheettracer/internal/db"
	"sheettracer/internal/keyring"
	"sheettracer/internal/oauth"
	"sheettracer/internal/services"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// Wails uses Go's `embed` package to embed the frontend files into the binary.
// Any files in the frontend/dist folder will be embedded into the binary and
// made available to the frontend.
// See https://pkg.go.dev/embed for more information.

//go:embed all:frontend/dist
var assets embed.FS

// main function serves as the application's entry point. It initializes the application and creates a window.
func main() {
	// Dev builds read a gitignored .env for SHEETTRACER_GOOGLE_*; prod builds no-op.
	oauth.LoadDotEnv()

	confDir, err := os.UserConfigDir()
	if err != nil {
		panic(err)
	}

	path := filepath.Join(confDir, "SheetTracer", "sheetTracer.db")
	if _, err = database.Open(path); err != nil {
		panic(err)
	}

	kr, err := keyring.NewStore("SheetTracer", filepath.Join(confDir, "SheetTracer", "keyring"))
	if err != nil {
		panic(err)
	}

	// _ = db
	// Create a new Wails application by providing the necessary options.
	// Variables 'Name' and 'Description' are for application metadata.
	// 'Assets' configures the asset server with the 'FS' variable pointing to the frontend files.
	// 'Bind' is a list of Go struct instances. The frontend has access to the methods of these instances.
	// 'Mac' options tailor the application when running on macOS.
	app := application.New(application.Options{
		Name:        "SheetTracer",
		Description: "Map IMPORTRANGE dependencies across your Google Sheets.",
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	oauthService := services.NewOAuthService(
		oauth.New(clientID(), clientSecret()),
		oauth.NewKeyringStore(kr),
		app.Browser.OpenURL,
		func(event string, data any) { app.Event.Emit(event, data) },
	)
	app.RegisterService(application.NewService(oauthService))

	menu := createMenu(app)
	app.Menu.Set(menu)

	// Create a new window with the necessary options.
	// 'Title' is the title of the window.
	// 'Mac' options tailor the window when running on macOS.
	// 'BackgroundColour' is the background colour of the window.
	// 'URL' is the URL that will be loaded into the webview.
	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title: "SheetTracer",
		// Window sized to the golden ratio (1000 / 618 ≈ 1.618).
		Width:  1000,
		Height: 618,
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundColour: application.NewRGB(6, 7, 15),
		URL:              "/",
	})

	// Run the application. This blocks until the application has been exited.
	err = app.Run()

	// If an error occurred while running the application, log it and exit.
	if err != nil {
		log.Fatal(err)
	}
}

func createMenu(app *application.App) *application.Menu {
	menu := app.NewMenu()

	if runtime.GOOS == "darwin" {
		menu.AddRole(application.AppMenu)
	}

	if runtime.GOOS != "darwin" {
		fileMenu := menu.AddSubmenu("File")
		fileMenu.Add("Exit").SetAccelerator("Alt+F4").OnClick(func(ctx *application.Context) {
			app.Quit()
		})
	}

	helpMenu := menu.AddSubmenu("Help")
	helpMenu.Add("Documentation").OnClick(func(ctx *application.Context) {
		app.Browser.OpenURL("https://github.com/farrelaby/sheettracer/tree/main/docs")
	})
	helpMenu.Add("Report Issue").OnClick(func(ctx *application.Context) {
		app.Browser.OpenURL("https://github.com/farrelaby/sheettracer/issues")
	})
	helpMenu.AddSeparator()
	helpMenu.Add("About SheetTracer").OnClick(func(ctx *application.Context) {
		app.Dialog.Info().
			SetTitle("About SheetTracer").
			SetMessage("SheetTracer\n\nMap IMPORTRANGE dependencies across your Google Sheets.").
			Show()
	})

	return menu
}

func clientID() string {
	if v := os.Getenv("SHEETTRACER_GOOGLE_CLIENT_ID"); v != "" {
		return v
	}
	if oauth.DefaultClientID != "" {
		return oauth.DefaultClientID
	}
	log.Fatal("no Google OAuth client configured: set SHEETTRACER_GOOGLE_CLIENT_ID (dev .env) or build with -tags production")
	return ""
}

func clientSecret() string {
	if v := os.Getenv("SHEETTRACER_GOOGLE_CLIENT_SECRET"); v != "" {
		return v
	}
	if oauth.DefaultClientSecret != "" {
		return oauth.DefaultClientSecret
	}
	log.Fatal("no Google OAuth client configured: set SHEETTRACER_GOOGLE_CLIENT_SECRET (dev .env) or build with -tags production")
	return ""
}
