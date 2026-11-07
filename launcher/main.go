// Last Light – Starter: öffnet das Spiel in einem eigenen Vollbild-Fenster (ohne Browser-Leisten).
package main

import (
	"embed"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"unsafe"
)

//go:embed game/index.html
var content embed.FS

func msgBox(text string) {
	u := syscall.NewLazyDLL("user32.dll").NewProc("MessageBoxW")
	t, _ := syscall.UTF16PtrFromString(text)
	c, _ := syscall.UTF16PtrFromString("Last Light")
	u.Call(0, uintptr(unsafe.Pointer(t)), uintptr(unsafe.Pointer(c)), 0x10)
}

func exists(p string) bool { _, err := os.Stat(p); return err == nil }

func main() {
	// Fester Port, damit Rekorde (localStorage) erhalten bleiben
	ln, err := net.Listen("tcp", "127.0.0.1:47831")
	if err != nil {
		ln, err = net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			msgBox("Konnte den Spiel-Server nicht starten.")
			return
		}
	}
	sub, _ := fs.Sub(content, "game")
	go http.Serve(ln, http.FileServer(http.FS(sub)))
	url := "http://" + ln.Addr().String() + "/index.html?app=1"

	base := os.Getenv("LOCALAPPDATA")
	if base == "" {
		base = os.TempDir()
	}
	profile := filepath.Join(base, "LastLight", "profile")
	os.MkdirAll(profile, 0755)

	pf, pf86, la := os.Getenv("ProgramFiles"), os.Getenv("ProgramFiles(x86)"), os.Getenv("LOCALAPPDATA")
	candidates := []string{
		filepath.Join(pf86, `Microsoft\Edge\Application\msedge.exe`),
		filepath.Join(pf, `Microsoft\Edge\Application\msedge.exe`),
		filepath.Join(pf, `Google\Chrome\Application\chrome.exe`),
		filepath.Join(pf86, `Google\Chrome\Application\chrome.exe`),
		filepath.Join(la, `Google\Chrome\Application\chrome.exe`),
		filepath.Join(pf, `BraveSoftware\Brave-Browser\Application\brave.exe`),
	}
	args := []string{
		"--app=" + url,
		"--start-fullscreen",
		"--user-data-dir=" + profile,
		"--no-first-run",
		"--no-default-browser-check",
		"--disable-features=Translate,msEdgeSidebarV2",
		"--autoplay-policy=no-user-gesture-required",
		"--overscroll-history-navigation=0",
	}
	for _, b := range candidates {
		if b != "" && exists(b) {
			cmd := exec.Command(b, args...)
			if err := cmd.Start(); err == nil {
				cmd.Wait() // Programm endet, wenn das Spielfenster geschlossen wird
				return
			}
		}
	}
	msgBox("Microsoft Edge oder Google Chrome wurde nicht gefunden.\nBitte installiere Edge oder Chrome, um Last Light zu spielen.")
}
