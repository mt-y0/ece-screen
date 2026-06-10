package video

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/gen2brain/go-mpv"
)

var Player *mpv.Mpv

func MpvInit() *mpv.Mpv {
	m := mpv.New()
	_ = m.SetOptionString("background-color", "0.0/0.0/0.0")
	_ = m.SetOptionString("idle", "yes")
	_ = m.SetOptionString("force-window", "yes")
	_ = m.SetPropertyString("fullscreen", "yes")
	err := m.Initialize()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize MPV: %v", err))
	}

	time.Sleep(time.Millisecond * 400)

	return m
}

// Resolution returns the width and height of the first video stream in the file using ffprobe.
func Resolution(path string) (int32, int32, error) {
	out, err := exec.Command("ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height",
		"-of", "csv=p=0",
		path,
	).Output()
	if err != nil {
		return 0, 0, err
	}
	parts := strings.SplitN(strings.TrimSpace(string(out)), ",", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("unexpected ffprobe output: %q", out)
	}
	w, err1 := strconv.Atoi(parts[0])
	h, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		return 0, 0, fmt.Errorf("could not parse ffprobe output: %q", out)
	}
	return int32(w), int32(h), nil
}

// Unload stops the current file and frees its resources while keeping the window open.
func Unload() {
	_ = Player.Command([]string{"stop"})
}

// IsIdle reports whether MPV has finished playback and is now idle.
func IsIdle() bool {
	val, err := Player.GetProperty("idle-active", mpv.FormatFlag)
	if err != nil {
		return false
	}
	b, ok := val.(bool)
	return ok && b
}
