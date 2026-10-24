package main

import (
	_ "embed"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"mt-y0/ECE-screen/internal/components"
	"mt-y0/ECE-screen/internal/db"
	"mt-y0/ECE-screen/lib/emoji"
	lib "mt-y0/ECE-screen/lib/fonts"
	"mt-y0/ECE-screen/lib/video"

	rl "github.com/gen2brain/raylib-go/raylib"
)

//go:embed resources/fonts/notokufi.ttf
var arabicFontData []byte

// resourceDir returns rel if it exists relative to the CWD, otherwise relative
// to the executable's directory (so the binary also works when launched from
// elsewhere, e.g. by systemd). Falls back to rel unchanged.
func resourceDir(rel string) string {
	if _, err := os.Stat(rel); err == nil {
		return rel
	}
	if exe, err := os.Executable(); err == nil {
		if p := filepath.Join(filepath.Dir(exe), rel); dirExists(p) {
			return p
		}
	}
	return rel
}

func dirExists(p string) bool {
	st, err := os.Stat(p)
	return err == nil && st.IsDir()
}

func main() {
	pb := db.InitPocketBase()
	video.Player = video.MpvInit()
	defer video.Player.TerminateDestroy()
	// raylib logs every loaded file at INFO (~3000 lines for the emoji set);
	// show warnings and errors only unless RL_LOG=info|debug is set.
	logLevel := rl.LogWarning
	switch os.Getenv("RL_LOG") {
	case "info":
		logLevel = rl.LogInfo
	case "debug":
		logLevel = rl.LogDebug
	}
	rl.SetTraceLogLevel(logLevel)
	rl.SetConfigFlags(rl.FlagWindowTransparent)
	rl.InitWindow(0, 0, "ECE screen")
	rl.HideCursor()
	targetFPS := 60
	if v := os.Getenv("FPS"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			targetFPS = parsed
		}
	}
	rl.SetTargetFPS(int32(targetFPS))
	defer rl.CloseWindow()

	font := lib.LoadFont(arabicFontData)

	// Emoji atlas — loads from resources/fonts/twemoji/ (72×72 Twemoji PNGs).
	// Inert if the directory is missing; no nil-checks needed downstream.
	emojiAtlas := emoji.Load(resourceDir("resources/fonts/twemoji"))
	defer emojiAtlas.Unload()

	screenSize := [2]int32{
		int32(rl.GetScreenWidth()),
		int32(rl.GetScreenHeight()),
	}
	zoom := 1.0
	if v := os.Getenv("ZOOM"); v != "" {
		if parsed, err := strconv.ParseFloat(v, 64); err == nil && parsed > 0 {
			zoom = parsed
		} else {
			log.Printf("ignoring invalid ZOOM=%q (need a number > 0); using 1.0", v)
		}
	}

	blue := rl.NewColor(0, 86, 179, 255)
	lightBlue := rl.NewColor(231, 241, 255, 255)
	bgColor := rl.NewColor(248, 249, 250, 255)

	tableLayout := components.GetTableLayout(screenSize, zoom)
	defer tableLayout.CleanUp()
	if err := tableLayout.GetData(pb); err != nil {
		log.Printf("table data: %v", err)
	}
	tableLayout.Watch(pb)

	tickerBarLayout := components.GetTickerBarLayout(screenSize, zoom)
	if err := tickerBarLayout.GetData(pb); err != nil {
		log.Printf("ticker data: %v", err)
	}
	tickerBarLayout.Watch(pb)

	professorsBoard := components.GetProfessorsBoardLayout(screenSize, zoom)
	defer professorsBoard.CleanUp()

	secMedia := components.GetSecMediaLayout(screenSize, zoom)
	defer secMedia.CleanUp()
	if err := secMedia.GetData(pb); err != nil {
		log.Printf("secmedia data: %v", err)
	}
	secMedia.Watch(pb)

	mainMedia := components.GetMainMediaLayout(screenSize, zoom)
	defer mainMedia.CleanUp()
	if err := mainMedia.GetData(pb); err != nil {
		log.Printf("mainmedia data: %v", err)
	}
	mainMedia.Watch(pb)

	if err := professorsBoard.GetData(pb); err != nil {
		log.Printf("professors data: %v", err)
	}
	professorsBoard.Watch(pb)

	noticeBoard := components.GetNoticeBoardLayout(screenSize, zoom)
	defer noticeBoard.CleanUp()
	if err := noticeBoard.GetData(pb); err != nil {
		log.Printf("notice data: %v", err)
	}
	noticeBoard.Watch(pb)

	for !rl.WindowShouldClose() {
		if rl.IsKeyPressed(rl.KeyX) {
			break
		}

		if mainMedia.IsVideo && video.IsIdle() {
			mainMedia.IsVideo = false
			video.Unload()
		}

		rl.BeginDrawing()
		rl.ClearBackground(bgColor)

		tableLayout.Draw(font, zoom, blue, lightBlue)
		professorsBoard.Draw(font, emojiAtlas, zoom, blue, lightBlue)
		secMedia.Draw(font, emojiAtlas, zoom, blue, lightBlue)
		mainMedia.Draw(font, emojiAtlas, zoom, blue, lightBlue)
		tickerBarLayout.Draw(screenSize, font, emojiAtlas, blue)
		noticeBoard.Draw(font, emojiAtlas, zoom, blue, lightBlue)

		if mainMedia.IsVideo {
			rl.BeginBlendMode(rl.BlendCustomSeparate)
			rl.SetBlendFactorsSeparate(
				int32(0), int32(0),
				int32(0), int32(0),
				int32(0x8006), int32(0x8006),
			)

			rl.DrawRectangleRounded(mainMedia.BgRect, 0.03, 8, rl.White)

			rl.EndBlendMode()
		}

		rl.EndDrawing()
	}
}
