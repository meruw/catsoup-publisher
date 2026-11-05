package main

import (
	"embed"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync/atomic"

	"catsoup/publisher/internal/delivery"
	"catsoup/publisher/internal/drafts"
	"catsoup/publisher/internal/uploads"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

//go:embed all:frontend/dist
var bundled embed.FS

//go:embed frontend/public/brand/catsoup-rounded.png
var appIcon []byte

func findRoot() (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", err
	}
	start := filepath.Dir(executable)
	for {
		if _, err := os.Stat(filepath.Join(start, "contracts", "delivery", "delivery.py")); err == nil {
			return start, nil
		}
		next := filepath.Dir(start)
		if next == start {
			break
		}
		start = next
	}
	return "", errors.New("CATSOUP workspace not found; launch with --root PATH")
}
func main() {
	rootFlag := flag.String("root", "", "CATSOUP workspace path")
	check := flag.Bool("check", false, "validate delivery catalog without opening a window")
	flag.Parse()
	root := *rootFlag
	if root == "" {
		var err error
		root, err = findRoot()
		if err != nil {
			log.Fatal(err)
		}
	}
	root, err := filepath.Abs(root)
	if err != nil {
		log.Fatal(err)
	}
	reader, err := delivery.New(root, delivery.PythonValidator{Root: root})
	if err != nil {
		log.Fatal(err)
	}
	if *check {
		entries, err := reader.List()
		if err != nil {
			log.Fatal(err)
		}
		failed := false
		for _, e := range entries {
			fmt.Printf("%s valid=%t %s\n", e.ID, e.Valid, e.Problem)
			failed = failed || !e.Valid
		}
		if failed {
			os.Exit(1)
		}
		return
	}
	assets, err := fs.Sub(bundled, "frontend/dist")
	if err != nil {
		log.Fatal(err)
	}
	store, err := drafts.Open(filepath.Join(root, "data", "publisher", "drafts.sqlite"))
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()
	draftService := &DraftService{store: store, reader: reader}
	uploadService := &UploadService{store: store, reader: reader, client: uploads.New()}
	app := application.New(application.Options{Name: "CATsoup Publisher", Description: "Local release preparation",
		Icon:     appIcon,
		Services: []application.Service{application.NewService(&PublisherService{reader: reader}), application.NewService(draftService), application.NewService(uploadService)},
		Assets:   application.AssetOptions{Handler: assetHandler(reader, assets)}})
	window := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title: "CATsoup Publisher", Width: 1200, Height: 840, MinWidth: 820, MinHeight: 650, URL: "/", Frameless: true,
		BackgroundColour: application.NewRGB(255, 252, 247), DevToolsEnabled: false,
		Permissions: map[application.PermissionType]application.Permission{
			application.PermissionCamera:        application.PermissionDeny,
			application.PermissionMicrophone:    application.PermissionDeny,
			application.PermissionGeolocation:   application.PermissionDeny,
			application.PermissionNotifications: application.PermissionDeny,
			application.PermissionClipboardRead: application.PermissionDeny,
		},
	})
	draftService.copyText = app.Clipboard.SetText
	uploadService.openBrowser = app.Browser.OpenURL
	draftService.reveal = func(path string) error { return app.Env.OpenFileManager(path, true) }
	var guarded, closing atomic.Bool
	draftService.enableCloseGuard = func() { guarded.Store(true) }
	draftService.confirmClose = func() { closing.Store(true); window.Close() }
	window.RegisterHook(events.Common.WindowClosing, func(event *application.WindowEvent) {
		if guarded.Load() && !closing.Load() {
			event.Cancel()
			app.Event.Emit("publisher:close-requested")
		}
	})
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}

func assetHandler(reader *delivery.Reader, assets fs.FS) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/media", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" && r.Method != "HEAD" {
			http.Error(w, "Method not allowed", 405)
			return
		}
		path, err := reader.Media(r.URL.Query().Get("episode"), r.URL.Query().Get("video"))
		if err != nil {
			http.Error(w, "Media unavailable: delivery must be valid", 400)
			return
		}
		file, err := os.Open(path)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer file.Close()
		stat, err := file.Stat()
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "video/mp4")
		w.Header().Set("Cache-Control", "no-store")
		http.ServeContent(w, r, filepath.Base(path), stat.ModTime(), file)
	})
	mux.Handle("/", http.FileServer(http.FS(assets)))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; media-src 'self'; connect-src 'self'; frame-src 'none'; object-src 'none'; base-uri 'none'; form-action 'none'")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		mux.ServeHTTP(w, r)
	})
}
